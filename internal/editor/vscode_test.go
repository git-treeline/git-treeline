package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVSCode_CreatesSettings(t *testing.T) {
	dir := t.TempDir()
	s := VSCodeSettings{
		Title: "test-project :3010 (main)",
		Color: "#1a5276",
	}

	target, err := WriteVSCode(dir, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target == "" {
		t.Fatal("expected target path")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read settings: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if settings["window.title"] != "test-project :3010 (main)" {
		t.Errorf("unexpected title: %v", settings["window.title"])
	}

	colors, ok := settings["workbench.colorCustomizations"].(map[string]any)
	if !ok {
		t.Fatal("expected colorCustomizations")
	}
	if colors["titleBar.activeBackground"] != "#1a5276" {
		t.Errorf("unexpected color: %v", colors["titleBar.activeBackground"])
	}
}

func TestWriteVSCode_WithTheme(t *testing.T) {
	dir := t.TempDir()
	s := VSCodeSettings{
		Title: "test",
		Theme: "Monokai",
	}

	target, err := WriteVSCode(dir, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(target)
	var settings map[string]any
	_ = json.Unmarshal(data, &settings)

	if settings["workbench.colorTheme"] != "Monokai" {
		t.Errorf("unexpected theme: %v", settings["workbench.colorTheme"])
	}
}

func TestWriteVSCode_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	vscodeDir := filepath.Join(dir, ".vscode")
	_ = os.MkdirAll(vscodeDir, 0o755)

	existing := map[string]any{
		"editor.fontSize": 14,
		"go.gopath":       "/home/user/go",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	_ = os.WriteFile(filepath.Join(vscodeDir, "settings.json"), data, 0o644)

	s := VSCodeSettings{Title: "new-title"}
	target, err := WriteVSCode(dir, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, _ := os.ReadFile(target)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)

	if result["window.title"] != "new-title" {
		t.Errorf("expected new title, got %v", result["window.title"])
	}
	if result["editor.fontSize"] != float64(14) {
		t.Errorf("expected preserved fontSize, got %v", result["editor.fontSize"])
	}
}

func TestWriteVSCode_UsesWorkspaceFile(t *testing.T) {
	parentDir := t.TempDir()
	projectDir := filepath.Join(parentDir, "my-project")
	_ = os.MkdirAll(projectDir, 0o755)

	ws := map[string]any{
		"folders":  []any{map[string]any{"path": "my-project"}},
		"settings": map[string]any{},
	}
	wsData, _ := json.MarshalIndent(ws, "", "\t")
	wsPath := filepath.Join(parentDir, "my-project.code-workspace")
	_ = os.WriteFile(wsPath, wsData, 0o644)

	s := VSCodeSettings{Title: "workspace-test", Color: "#7b241c"}
	target, err := WriteVSCode(projectDir, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target != wsPath {
		t.Errorf("expected workspace file %s, got %s", wsPath, target)
	}

	raw, _ := os.ReadFile(wsPath)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	settings, _ := result["settings"].(map[string]any)
	if settings["window.title"] != "workspace-test" {
		t.Errorf("expected title in workspace file, got %v", settings["window.title"])
	}
}

func TestWriteVSCode_ColorOnly(t *testing.T) {
	dir := t.TempDir()
	s := VSCodeSettings{Color: "#196f3d"}

	target, err := WriteVSCode(dir, s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(target)
	var settings map[string]any
	_ = json.Unmarshal(data, &settings)

	if _, ok := settings["window.title"]; ok {
		t.Error("expected no window.title when Title is empty")
	}
	colors, ok := settings["workbench.colorCustomizations"].(map[string]any)
	if !ok {
		t.Fatal("expected colorCustomizations for color-only settings")
	}
	if colors["titleBar.activeBackground"] != "#196f3d" {
		t.Errorf("unexpected bg: %v", colors["titleBar.activeBackground"])
	}
}

func TestBuildSettings_Empty(t *testing.T) {
	s := VSCodeSettings{}
	settings := buildSettings(s)
	if len(settings) != 0 {
		t.Errorf("expected empty settings map, got %v", settings)
	}
}

func TestFindWorkspaceFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	if ws := findWorkspaceFile(dir); ws != "" {
		t.Errorf("expected no workspace, got %s", ws)
	}
}
