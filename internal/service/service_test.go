package service

import (
	"strings"
	"testing"
)

func TestGeneratePlist(t *testing.T) {
	content, err := GeneratePlist("/usr/local/bin/gtl")
	if err != nil {
		t.Fatalf("GeneratePlist failed: %v", err)
	}

	checks := []string{
		"<string>dev.treeline.router</string>",
		"<string>/usr/local/bin/gtl</string>",
		"<string>serve</string>",
		"<string>run</string>",
		"<key>RunAtLoad</key>",
		"<key>KeepAlive</key>",
		"router.log",
		"router.err",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("plist missing %q", check)
		}
	}
}

func TestGenerateUnit(t *testing.T) {
	content, err := GenerateUnit("/usr/local/bin/gtl")
	if err != nil {
		t.Fatalf("GenerateUnit failed: %v", err)
	}

	checks := []string{
		"ExecStart=/usr/local/bin/gtl serve run",
		"Restart=always",
		"WantedBy=default.target",
		"git-treeline subdomain router",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("unit missing %q", check)
		}
	}
}
