package config

import (
	"testing"
)

func TestDeepMerge_OverrideWins(t *testing.T) {
	base := map[string]any{"a": "1", "b": "2"}
	over := map[string]any{"a": "X"}
	result := DeepMerge(base, over)
	if result["a"] != "X" {
		t.Errorf("expected override value X, got %v", result["a"])
	}
	if result["b"] != "2" {
		t.Errorf("expected base value 2, got %v", result["b"])
	}
}

func TestDeepMerge_NestedMaps(t *testing.T) {
	base := map[string]any{
		"port": map[string]any{"base": 3000, "increment": 10},
	}
	over := map[string]any{
		"port": map[string]any{"base": 4000},
	}
	result := DeepMerge(base, over)
	port := result["port"].(map[string]any)
	if port["base"] != 4000 {
		t.Errorf("expected 4000, got %v", port["base"])
	}
	if port["increment"] != 10 {
		t.Errorf("expected 10 from base, got %v", port["increment"])
	}
}

func TestDeepMerge_OverrideReplacesNonMap(t *testing.T) {
	base := map[string]any{"key": "string_value"}
	over := map[string]any{"key": map[string]any{"nested": true}}
	result := DeepMerge(base, over)
	if _, ok := result["key"].(map[string]any); !ok {
		t.Errorf("expected map, got %T", result["key"])
	}
}

func TestDig(t *testing.T) {
	m := map[string]any{
		"port": map[string]any{
			"base": 3000,
		},
	}
	if v := Dig(m, "port", "base"); v != 3000 {
		t.Errorf("expected 3000, got %v", v)
	}
	if v := Dig(m, "port", "missing"); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
	if v := Dig(m, "nope"); v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}
