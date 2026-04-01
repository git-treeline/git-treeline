package platform

import (
	"strings"
	"testing"
)

func TestConfigDir_NotEmpty(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
	if !strings.HasSuffix(dir, "git-treeline") {
		t.Errorf("expected path ending in git-treeline, got %s", dir)
	}
}

func TestConfigFile_EndsWithJSON(t *testing.T) {
	f := ConfigFile()
	if !strings.HasSuffix(f, "config.json") {
		t.Errorf("expected config.json, got %s", f)
	}
}

func TestRegistryFile_EndsWithJSON(t *testing.T) {
	f := RegistryFile()
	if !strings.HasSuffix(f, "registry.json") {
		t.Errorf("expected registry.json, got %s", f)
	}
}
