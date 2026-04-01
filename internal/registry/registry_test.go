package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestRegistry(t *testing.T) *Registry {
	t.Helper()
	dir := t.TempDir()
	return New(filepath.Join(dir, "registry.json"))
}

func TestRegistry_AllocateAndFind(t *testing.T) {
	reg := newTestRegistry(t)

	entry := Allocation{
		"project":       "salt",
		"worktree":      "/tmp/wt/branch-a",
		"worktree_name": "branch-a",
		"port":          float64(3010),
		"ports":         []any{float64(3010), float64(3011)},
		"database":      "salt_development_branch_a",
	}

	if err := reg.Allocate(entry); err != nil {
		t.Fatal(err)
	}

	found := reg.Find("/tmp/wt/branch-a")
	if found == nil {
		t.Fatal("expected allocation, got nil")
	}
	if found["worktree_name"] != "branch-a" {
		t.Errorf("expected branch-a, got %v", found["worktree_name"])
	}
}

func TestRegistry_UsedPorts_MultiPort(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Allocate(Allocation{
		"project":  "a",
		"worktree": "/wt/a",
		"ports":    []any{float64(3010), float64(3011)},
	})
	reg.Allocate(Allocation{
		"project":  "b",
		"worktree": "/wt/b",
		"port":     float64(3020),
	})

	ports := reg.UsedPorts()
	if len(ports) != 3 {
		t.Errorf("expected 3 ports, got %d: %v", len(ports), ports)
	}
}

func TestRegistry_UsedPorts_BackwardCompat(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Allocate(Allocation{
		"project":  "old",
		"worktree": "/wt/old",
		"port":     float64(3010),
	})

	ports := reg.UsedPorts()
	if len(ports) != 1 || ports[0] != 3010 {
		t.Errorf("expected [3010], got %v", ports)
	}
}

func TestRegistry_Release(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Allocate(Allocation{"worktree": "/wt/x"})
	reg.Allocate(Allocation{"worktree": "/wt/y"})

	removed, _ := reg.Release("/wt/x")
	if !removed {
		t.Error("expected removal")
	}
	if found := reg.Find("/wt/x"); found != nil {
		t.Error("expected nil after release")
	}
	if found := reg.Find("/wt/y"); found == nil {
		t.Error("expected /wt/y to still exist")
	}
}

func TestRegistry_Prune(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "worktree-a")
	os.MkdirAll(existing, 0o755)

	reg := newTestRegistry(t)
	reg.Allocate(Allocation{"worktree": existing})
	reg.Allocate(Allocation{"worktree": "/does/not/exist"})

	count, _ := reg.Prune()
	if count != 1 {
		t.Errorf("expected 1 pruned, got %d", count)
	}

	if len(reg.Allocations()) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(reg.Allocations()))
	}
}

func TestRegistry_AllocateReplaces(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Allocate(Allocation{"worktree": "/wt/a", "port": float64(3010)})
	reg.Allocate(Allocation{"worktree": "/wt/a", "port": float64(3020)})

	allocs := reg.Allocations()
	if len(allocs) != 1 {
		t.Errorf("expected 1 allocation after re-allocate, got %d", len(allocs))
	}
	if allocs[0]["port"] != float64(3020) {
		t.Errorf("expected updated port 3020, got %v", allocs[0]["port"])
	}
}

func TestRegistry_PruneStale_RemovesMissingDirs(t *testing.T) {
	reg := newTestRegistry(t)
	reg.Allocate(Allocation{"worktree": "/does/not/exist/at/all"})

	count, err := reg.PruneStale()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("expected 1 pruned, got %d", count)
	}
}
