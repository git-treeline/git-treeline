package allocator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/registry"
)

func testAllocator(t *testing.T, portsNeeded int, yamlExtra string) (*Allocator, *registry.Registry) {
	t.Helper()

	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	reg := registry.New(regPath)

	confPath := filepath.Join(dir, "config.json")
	os.WriteFile(confPath, []byte(`{"port":{"base":3000,"increment":10},"redis":{"strategy":"prefixed","url":"redis://localhost:6379"}}`), 0o644)
	uc := config.LoadUserConfig(confPath)

	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0o755)
	yml := "project: test\n"
	if portsNeeded > 0 {
		yml += "ports_needed: " + itoa(portsNeeded) + "\n"
	}
	yml += "database:\n  adapter: postgresql\n  template: test_dev\n  pattern: \"{template}_{worktree}\"\n"
	yml += yamlExtra
	os.WriteFile(filepath.Join(projDir, ".treeline.yml"), []byte(yml), 0o644)
	pc := config.LoadProjectConfig(projDir)

	return New(uc, pc, reg), reg
}

func itoa(n int) string {
	return string(rune('0') + rune(n))
}

func TestAllocate_SinglePort(t *testing.T) {
	al, _ := testAllocator(t, 1, "")
	alloc, err := al.Allocate("/wt/branch-a", "branch-a")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.Port != 3010 {
		t.Errorf("expected 3010, got %d", alloc.Port)
	}
	if len(alloc.Ports) != 1 {
		t.Errorf("expected 1 port, got %d", len(alloc.Ports))
	}
}

func TestAllocate_MultiPort(t *testing.T) {
	al, _ := testAllocator(t, 2, "")
	alloc, err := al.Allocate("/wt/branch-a", "branch-a")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.Port != 3010 {
		t.Errorf("expected 3010, got %d", alloc.Port)
	}
	if len(alloc.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(alloc.Ports))
	}
	if alloc.Ports[1] != 3011 {
		t.Errorf("expected second port 3011, got %d", alloc.Ports[1])
	}
}

func TestAllocate_SkipsUsedPorts(t *testing.T) {
	al, reg := testAllocator(t, 1, "")
	reg.Allocate(registry.Allocation{
		"worktree": "/wt/existing",
		"port":     float64(3010),
		"ports":    []any{float64(3010)},
	})

	alloc, err := al.Allocate("/wt/new", "new")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.Port != 3020 {
		t.Errorf("expected 3020 (skipped 3010), got %d", alloc.Port)
	}
}

func TestAllocate_MultiPort_NonOverlapping(t *testing.T) {
	al, reg := testAllocator(t, 2, "")
	reg.Allocate(registry.Allocation{
		"worktree": "/wt/existing",
		"port":     float64(3010),
		"ports":    []any{float64(3010), float64(3011)},
	})

	alloc, err := al.Allocate("/wt/new", "new")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.Port != 3020 {
		t.Errorf("expected 3020, got %d", alloc.Port)
	}
}

func TestAllocate_PortsNeededExceedsIncrement(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	reg := registry.New(regPath)

	confPath := filepath.Join(dir, "config.json")
	os.WriteFile(confPath, []byte(`{"port":{"base":3000,"increment":2}}`), 0o644)
	uc := config.LoadUserConfig(confPath)

	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0o755)
	os.WriteFile(filepath.Join(projDir, ".treeline.yml"), []byte("project: test\nports_needed: 5\n"), 0o644)
	pc := config.LoadProjectConfig(projDir)

	al := New(uc, pc, reg)
	_, err := al.Allocate("/wt/x", "x")
	if err == nil {
		t.Fatal("expected error when ports_needed > increment")
	}
}

func TestAllocate_DatabaseName(t *testing.T) {
	al, _ := testAllocator(t, 1, "")
	alloc, err := al.Allocate("/wt/feature-branch", "feature-branch")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.Database != "test_dev_feature_branch" {
		t.Errorf("expected test_dev_feature_branch, got %s", alloc.Database)
	}
}

func TestAllocate_RedisPrefix(t *testing.T) {
	al, _ := testAllocator(t, 1, "")
	alloc, err := al.Allocate("/wt/my-branch", "my-branch")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.RedisPrefix != "test:my-branch" {
		t.Errorf("expected test:my-branch, got %s", alloc.RedisPrefix)
	}
}

func TestAllocate_RedisDatabaseStrategy(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.json")
	reg := registry.New(regPath)

	confPath := filepath.Join(dir, "config.json")
	os.WriteFile(confPath, []byte(`{"port":{"base":3000,"increment":10},"redis":{"strategy":"database","url":"redis://localhost:6379"}}`), 0o644)
	uc := config.LoadUserConfig(confPath)

	projDir := filepath.Join(dir, "project")
	os.MkdirAll(projDir, 0o755)
	os.WriteFile(filepath.Join(projDir, ".treeline.yml"), []byte("project: test\ndatabase:\n  adapter: postgresql\n  template: test_dev\n  pattern: \"{template}_{worktree}\"\n"), 0o644)
	pc := config.LoadProjectConfig(projDir)

	al := New(uc, pc, reg)
	alloc, err := al.Allocate("/wt/a", "a")
	if err != nil {
		t.Fatal(err)
	}
	if alloc.RedisDB != 1 {
		t.Errorf("expected redis db 1, got %d", alloc.RedisDB)
	}
}

func TestToRegistryEntry_Format(t *testing.T) {
	alloc := &Allocation{
		Project:      "salt",
		Worktree:     "/wt/branch",
		WorktreeName: "branch",
		Port:         3010,
		Ports:        []int{3010, 3011},
		Database:     "salt_dev_branch",
		RedisPrefix:  "salt:branch",
	}

	entry := alloc.ToRegistryEntry()
	raw, _ := json.Marshal(entry)
	var parsed map[string]any
	json.Unmarshal(raw, &parsed)

	if parsed["project"] != "salt" {
		t.Errorf("expected salt, got %v", parsed["project"])
	}
	if ports, ok := parsed["ports"].([]any); ok {
		if len(ports) != 2 {
			t.Errorf("expected 2 ports, got %d", len(ports))
		}
	} else {
		t.Error("expected ports array")
	}
}
