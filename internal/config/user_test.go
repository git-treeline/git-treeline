package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserConfig_Defaults(t *testing.T) {
	uc := LoadUserConfig("/nonexistent/config.json")
	if uc.PortBase() != 3000 {
		t.Errorf("expected 3000, got %d", uc.PortBase())
	}
	if uc.PortIncrement() != 10 {
		t.Errorf("expected 10, got %d", uc.PortIncrement())
	}
	if uc.RedisStrategy() != "prefixed" {
		t.Errorf("expected prefixed, got %s", uc.RedisStrategy())
	}
	if uc.RedisURL() != "redis://localhost:6379" {
		t.Errorf("expected redis://localhost:6379, got %s", uc.RedisURL())
	}
}

func TestUserConfig_CustomValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{"port":{"base":4000,"increment":20}}`), 0o644)

	uc := LoadUserConfig(path)
	if uc.PortBase() != 4000 {
		t.Errorf("expected 4000, got %d", uc.PortBase())
	}
	if uc.PortIncrement() != 20 {
		t.Errorf("expected 20, got %d", uc.PortIncrement())
	}
	if uc.RedisStrategy() != "prefixed" {
		t.Errorf("expected prefixed default, got %s", uc.RedisStrategy())
	}
}

func TestUserConfig_Init(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")
	uc := LoadUserConfig(path)

	if uc.Exists() {
		t.Error("expected Exists() to be false before init")
	}
	if err := uc.Init(); err != nil {
		t.Fatal(err)
	}
	if !uc.Exists() {
		t.Error("expected Exists() to be true after init")
	}
}
