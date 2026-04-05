// Package service manages the git-treeline router as a system service.
// Supports macOS LaunchAgents and Linux systemd user units.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const launchLabel = "dev.treeline.router"
const systemdUnit = "git-treeline-router.service"

// Install writes a service definition and activates it.
// Returns the path to the written file.
func Install(gtlPath string, port int) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return installLaunchAgent(gtlPath, port)
	case "linux":
		return installSystemd(gtlPath, port)
	default:
		return "", fmt.Errorf("unsupported platform: %s (macOS and Linux only)", runtime.GOOS)
	}
}

// Uninstall stops the service and removes the definition file.
func Uninstall() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallLaunchAgent()
	case "linux":
		return uninstallSystemd()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// IsRunning checks if the service is currently active.
func IsRunning() bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("launchctl", "list", launchLabel).CombinedOutput()
		return err == nil && len(out) > 0
	case "linux":
		err := exec.Command("systemctl", "--user", "is-active", "--quiet", systemdUnit).Run()
		return err == nil
	default:
		return false
	}
}

// --- macOS LaunchAgent ---

var plistTemplate = template.Must(template.New("plist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{ .Label }}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{ .GtlPath }}</string>
		<string>serve</string>
		<string>run</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{ .LogDir }}/router.log</string>
	<key>StandardErrorPath</key>
	<string>{{ .LogDir }}/router.err</string>
</dict>
</plist>
`))

func PlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchLabel+".plist")
}

func logDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Logs", "git-treeline")
}

func installLaunchAgent(gtlPath string, _ int) (string, error) {
	path := PlistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(logDir(), 0o755); err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	err = plistTemplate.Execute(f, struct {
		Label   string
		GtlPath string
		LogDir  string
	}{
		Label:   launchLabel,
		GtlPath: gtlPath,
		LogDir:  logDir(),
	})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}

	_ = exec.Command("launchctl", "unload", path).Run()
	if err := exec.Command("launchctl", "load", path).Run(); err != nil {
		return path, fmt.Errorf("wrote plist but failed to load: %w", err)
	}
	return path, nil
}

func uninstallLaunchAgent() error {
	path := PlistPath()
	_ = exec.Command("launchctl", "unload", path).Run()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// --- Linux systemd user unit ---

var unitTemplate = template.Must(template.New("unit").Parse(`[Unit]
Description=git-treeline subdomain router

[Service]
ExecStart={{ .GtlPath }} serve run
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
`))

func UnitPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "systemd", "user", systemdUnit)
}

func installSystemd(gtlPath string, _ int) (string, error) {
	path := UnitPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	err = unitTemplate.Execute(f, struct{ GtlPath string }{GtlPath: gtlPath})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", err
	}

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", "--now", systemdUnit).Run(); err != nil {
		return path, fmt.Errorf("wrote unit but failed to enable: %w", err)
	}
	return path, nil
}

func uninstallSystemd() error {
	_ = exec.Command("systemctl", "--user", "disable", "--now", systemdUnit).Run()
	path := UnitPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	return nil
}

// GeneratePlist returns the plist XML content as a string (for testing).
func GeneratePlist(gtlPath string) (string, error) {
	var b strings.Builder
	err := plistTemplate.Execute(&b, struct {
		Label   string
		GtlPath string
		LogDir  string
	}{
		Label:   launchLabel,
		GtlPath: gtlPath,
		LogDir:  logDir(),
	})
	return b.String(), err
}

// GenerateUnit returns the systemd unit content as a string (for testing).
func GenerateUnit(gtlPath string) (string, error) {
	var b strings.Builder
	err := unitTemplate.Execute(&b, struct{ GtlPath string }{GtlPath: gtlPath})
	return b.String(), err
}
