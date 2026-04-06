package cmd

import (
	"strings"
	"testing"
)

func TestErrServeNotInstalled_ContainsGuidance(t *testing.T) {
	msg := errServeNotInstalled.Error()
	if !strings.Contains(msg, "gtl serve install") {
		t.Errorf("expected 'gtl serve install' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "gittreeline.com") {
		t.Errorf("expected docs URL in error, got: %s", msg)
	}
}

func TestSortedRouteKeys(t *testing.T) {
	routes := map[string]int{
		"myapp-feature":  3010,
		"myapp-main":     3000,
		"other-dev":      3020,
		"api-staging":    4000,
	}
	keys := sortedRouteKeys(routes)
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys, got %d", len(keys))
	}
	expected := []string{"api-staging", "myapp-feature", "myapp-main", "other-dev"}
	for i, want := range expected {
		if keys[i] != want {
			t.Errorf("keys[%d] = %q, want %q", i, keys[i], want)
		}
	}
}

func TestSortedRouteKeys_Empty(t *testing.T) {
	keys := sortedRouteKeys(map[string]int{})
	if len(keys) != 0 {
		t.Errorf("expected empty, got %v", keys)
	}
}

func TestSortedRouteKeys_Single(t *testing.T) {
	keys := sortedRouteKeys(map[string]int{"only": 3000})
	if len(keys) != 1 || keys[0] != "only" {
		t.Errorf("expected [only], got %v", keys)
	}
}
