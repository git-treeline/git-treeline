package interpolation

import (
	"testing"
)

func TestBuildRedisURL_WithDB(t *testing.T) {
	alloc := Allocation{"redis_db": float64(3)}
	url := BuildRedisURL("redis://localhost:6379", alloc)
	if url != "redis://localhost:6379/3" {
		t.Errorf("expected redis://localhost:6379/3, got %s", url)
	}
}

func TestBuildRedisURL_WithoutDB(t *testing.T) {
	alloc := Allocation{"redis_prefix": "salt:branch"}
	url := BuildRedisURL("redis://localhost:6379", alloc)
	if url != "redis://localhost:6379" {
		t.Errorf("expected redis://localhost:6379, got %s", url)
	}
}

func TestBuildRedisURL_TrailingSlash(t *testing.T) {
	alloc := Allocation{"redis_db": float64(5)}
	url := BuildRedisURL("redis://localhost:6379/", alloc)
	if url != "redis://localhost:6379/5" {
		t.Errorf("expected redis://localhost:6379/5, got %s", url)
	}
}

func TestInterpolate_BasicTokens(t *testing.T) {
	alloc := Allocation{
		"port":         float64(3010),
		"database":     "salt_dev_branch",
		"worktree_name": "branch",
		"redis_prefix": "salt:branch",
	}

	tests := []struct {
		pattern  string
		expected string
	}{
		{"{port}", "3010"},
		{"{database}", "salt_dev_branch"},
		{"http://localhost:{port}", "http://localhost:3010"},
		{"{project}/{worktree}", "salt/branch"},
	}

	for _, tt := range tests {
		result := Interpolate(tt.pattern, alloc, "redis://localhost:6379", "salt")
		if result != tt.expected {
			t.Errorf("Interpolate(%q) = %q, want %q", tt.pattern, result, tt.expected)
		}
	}
}

func TestInterpolate_MultiPort(t *testing.T) {
	alloc := Allocation{
		"port":  float64(3010),
		"ports": []any{float64(3010), float64(3011)},
	}

	result := Interpolate("{port_2}", alloc, "", "")
	if result != "3011" {
		t.Errorf("expected 3011, got %s", result)
	}
}

func TestInterpolate_PortN_NoPortsArray(t *testing.T) {
	alloc := Allocation{
		"port": float64(3010),
	}

	result := Interpolate("{port_2}", alloc, "", "")
	if result != "{port_2}" {
		t.Errorf("expected literal {port_2}, got %s", result)
	}
}

func TestInterpolate_IntPorts(t *testing.T) {
	alloc := Allocation{
		"port":  3010,
		"ports": []int{3010, 3011},
	}

	result := Interpolate("{port_1}", alloc, "", "")
	if result != "3010" {
		t.Errorf("expected 3010, got %s", result)
	}

	result = Interpolate("{port_2}", alloc, "", "")
	if result != "3011" {
		t.Errorf("expected 3011, got %s", result)
	}
}
