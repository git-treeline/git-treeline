package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-treeline/git-treeline/internal/registry"
)

func TestRouteKey(t *testing.T) {
	tests := []struct {
		project string
		branch  string
		want    string
	}{
		{"salt", "main", "salt-main"},
		{"salt", "feature/staff-reporting", "salt-staff-reporting"},
		{"salt", "chore/upgrade-rails", "salt-upgrade-rails"},
		{"salt", "bugfix/sentry-SALT-H92", "salt-sentry-salt-h92"},
		{"salt", "hotfix/urgent-fix", "salt-urgent-fix"},
		{"wildlife_platform", "feature/mcp-server", "wildlife-platform-mcp-server"},
		{"salt", "", "salt"},
		{"api-docs", "main", "api-docs-main"},
		{"salt", "feature/some--double-dash", "salt-some-double-dash"},
		{"MY_PROJECT", "Feature/BigThing", "my-project-bigthing"},
	}
	for _, tt := range tests {
		t.Run(tt.project+"/"+tt.branch, func(t *testing.T) {
			got := RouteKey(tt.project, tt.branch)
			if got != tt.want {
				t.Errorf("RouteKey(%q, %q) = %q, want %q", tt.project, tt.branch, got, tt.want)
			}
		})
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"localhost", ""},
		{"127.0.0.1", ""},
		{"::1", ""},
		{"salt-main.localhost", "salt-main"},
		{"salt-staff-reporting.localhost", "salt-staff-reporting"},
		{"SALT-MAIN.LOCALHOST", "salt-main"},
		{"something.example.com", ""},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			got := extractSubdomain(tt.host)
			if got != tt.want {
				t.Errorf("extractSubdomain(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func testRegistry(t *testing.T, allocs []registry.Allocation) *registry.Registry {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	data := registry.RegistryData{Version: 1, Allocations: allocs}
	raw, _ := json.Marshal(data)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return registry.New(path)
}

func TestRouterRoutesRequest(t *testing.T) {
	targetPort := freePort(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "hello from salt")
	})
	target := &http.Server{Addr: fmt.Sprintf(":%d", targetPort), Handler: mux}
	go func() { _ = target.ListenAndServe() }()
	defer func() { _ = target.Close() }()
	waitForPort(t, targetPort)

	reg := testRegistry(t, []registry.Allocation{
		{"project": "salt", "branch": "feature/staff-reporting", "port": float64(targetPort), "ports": []any{float64(targetPort)}, "worktree": "/tmp/salt-staff"},
	})

	router := NewRouter(0, reg)
	routes := router.Routes()
	if routes["salt-staff-reporting"] != targetPort {
		t.Fatalf("expected route salt-staff-reporting → %d, got %v", targetPort, routes)
	}

	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req.Host = "salt-staff-reporting.localhost:3000"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello from salt" {
		t.Errorf("expected 'hello from salt', got %q", string(body))
	}
}

func TestRouterStatusPage(t *testing.T) {
	reg := testRegistry(t, []registry.Allocation{
		{"project": "salt", "branch": "main", "port": float64(3001), "ports": []any{float64(3001)}, "worktree": "/tmp/salt-main"},
	})

	router := NewRouter(3000, reg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req.Host = "localhost:3000"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "salt-main") {
		t.Error("status page should list the salt-main route")
	}
}

func TestRouterNotFound(t *testing.T) {
	reg := testRegistry(t, []registry.Allocation{})
	router := NewRouter(3000, reg)
	ts := httptest.NewServer(router)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/", nil)
	req.Host = "nonexistent.localhost:3000"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestRouterRefreshPicksUpNewAllocations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	data := registry.RegistryData{Version: 1, Allocations: []registry.Allocation{}}
	raw, _ := json.Marshal(data)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	reg := registry.New(path)
	router := NewRouter(3000, reg)

	if len(router.Routes()) != 0 {
		t.Fatal("expected no routes initially")
	}

	data.Allocations = []registry.Allocation{
		{"project": "salt", "branch": "main", "port": float64(3001), "ports": []any{float64(3001)}, "worktree": "/tmp/salt"},
	}
	raw, _ = json.Marshal(data)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	router.refreshRoutes()
	routes := router.Routes()
	if routes["salt-main"] != 3001 {
		t.Errorf("expected salt-main → 3001 after refresh, got %v", routes)
	}
}
