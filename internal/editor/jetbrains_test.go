package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectJetBrains(t *testing.T) {
	dir := t.TempDir()
	if DetectJetBrains(dir) {
		t.Error("expected false without .idea")
	}

	_ = os.MkdirAll(filepath.Join(dir, ".idea"), 0o755)
	if !DetectJetBrains(dir) {
		t.Error("expected true with .idea")
	}
}

func TestWriteJetBrains_NoIdeaDir(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteJetBrains(dir, "#1a5276")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path without .idea, got %s", path)
	}
}

func TestWriteJetBrains_CreatesWorkspaceXML(t *testing.T) {
	dir := t.TempDir()
	ideaDir := filepath.Join(dir, ".idea")
	_ = os.MkdirAll(ideaDir, 0o755)

	path, err := WriteJetBrains(dir, "#1a5276")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected path")
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "1a5276") {
		t.Errorf("expected color in XML, got:\n%s", content)
	}
	if !strings.Contains(content, "ProjectColorInfo") {
		t.Errorf("expected ProjectColorInfo component, got:\n%s", content)
	}
}

func TestWriteJetBrains_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	ideaDir := filepath.Join(dir, ".idea")
	_ = os.MkdirAll(ideaDir, 0o755)

	existing := `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectColorInfo">
    <option name="customColor" value="ff0000" />
  </component>
</project>
`
	_ = os.WriteFile(filepath.Join(ideaDir, "workspace.xml"), []byte(existing), 0o644)

	path, err := WriteJetBrains(dir, "#1a5276")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "1a5276") {
		t.Errorf("expected updated color, got:\n%s", content)
	}
}

func TestWriteJetBrains_PreservesOtherComponents(t *testing.T) {
	dir := t.TempDir()
	ideaDir := filepath.Join(dir, ".idea")
	_ = os.MkdirAll(ideaDir, 0o755)

	existing := `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="VcsDirectoryMappings">
    <mapping directory="" vcs="Git" />
  </component>
  <component name="ProjectColorInfo">
    <option name="customColor" value="ff0000" />
  </component>
  <component name="ChangeListManager">
    <list default="true" name="Changes" />
  </component>
</project>
`
	_ = os.WriteFile(filepath.Join(ideaDir, "workspace.xml"), []byte(existing), 0o644)

	_, err := WriteJetBrains(dir, "#1a5276")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(ideaDir, "workspace.xml"))
	content := string(data)

	if !strings.Contains(content, "1a5276") {
		t.Errorf("expected updated color")
	}
	if !strings.Contains(content, "VcsDirectoryMappings") {
		t.Error("VcsDirectoryMappings component was destroyed")
	}
	if !strings.Contains(content, "ChangeListManager") {
		t.Error("ChangeListManager component was destroyed")
	}
	if strings.Contains(content, "ff0000") {
		t.Error("old color value should have been replaced")
	}
}

func TestWriteJetBrains_InsertsComponentWhenMissing(t *testing.T) {
	dir := t.TempDir()
	ideaDir := filepath.Join(dir, ".idea")
	_ = os.MkdirAll(ideaDir, 0o755)

	existing := `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="VcsDirectoryMappings">
    <mapping directory="" vcs="Git" />
  </component>
</project>
`
	_ = os.WriteFile(filepath.Join(ideaDir, "workspace.xml"), []byte(existing), 0o644)

	_, err := WriteJetBrains(dir, "#7d3c98")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(ideaDir, "workspace.xml"))
	content := string(data)

	if !strings.Contains(content, "7d3c98") {
		t.Error("expected color to be inserted")
	}
	if !strings.Contains(content, "VcsDirectoryMappings") {
		t.Error("existing component should be preserved")
	}
	if !strings.Contains(content, "ProjectColorInfo") {
		t.Error("expected ProjectColorInfo to be added")
	}
}
