package editor

import (
	"fmt"
	"strings"
	"testing"
)

func TestColorForBranch_Deterministic(t *testing.T) {
	c1 := ColorForBranch("feature/login")
	c2 := ColorForBranch("feature/login")
	if c1 != c2 {
		t.Errorf("same branch produced different colors: %s vs %s", c1, c2)
	}
}

func TestColorForBranch_DifferentBranches(t *testing.T) {
	c1 := ColorForBranch("main")
	c2 := ColorForBranch("feature/payment")
	if c1 == c2 {
		t.Logf("warning: different branches produced same color (possible but unlikely): %s", c1)
	}
}

func TestColorForBranch_EmptyBranch(t *testing.T) {
	c := ColorForBranch("")
	if !strings.HasPrefix(c, "#") {
		t.Errorf("expected hex color, got %s", c)
	}
}

func TestColorForBranch_ValidHex(t *testing.T) {
	branches := []string{"main", "develop", "feature/a", "fix/b", "release/1.0"}
	for _, b := range branches {
		c := ColorForBranch(b)
		if len(c) != 7 || c[0] != '#' {
			t.Errorf("branch %q: expected #rrggbb, got %s", b, c)
		}
	}
}

func TestDarkenHex(t *testing.T) {
	dark := DarkenHex("#1a5276")
	if dark == "#1a5276" {
		t.Error("darkened color should differ from original")
	}
	if !strings.HasPrefix(dark, "#") || len(dark) != 7 {
		t.Errorf("expected #rrggbb, got %s", dark)
	}
}

func TestForegroundForHex_Dark(t *testing.T) {
	fg := ForegroundForHex("#1a5276")
	if fg != "#ffffff" {
		t.Errorf("expected white text on dark bg, got %s", fg)
	}
}

func TestForegroundForHex_Light(t *testing.T) {
	fg := ForegroundForHex("#f0f0f0")
	if fg != "#1e1e1e" {
		t.Errorf("expected dark text on light bg, got %s", fg)
	}
}

func TestParseHex_InvalidInput(t *testing.T) {
	for _, input := range []string{"", "xyz", "#12", "not-a-color", "#gggggg"} {
		r, g, b := parseHex(input)
		if input == "" || input == "xyz" || input == "#12" || input == "not-a-color" {
			if r != 0 || g != 0 || b != 0 {
				t.Errorf("parseHex(%q): expected 0,0,0, got %d,%d,%d", input, r, g, b)
			}
		}
	}
}

func TestDarkenHex_Black(t *testing.T) {
	dark := DarkenHex("#000000")
	if dark != "#000000" {
		t.Errorf("darkening black should stay black, got %s", dark)
	}
}

func TestDarkenHex_White(t *testing.T) {
	dark := DarkenHex("#ffffff")
	if dark == "#ffffff" {
		t.Error("darkening white should produce a non-white color")
	}
	if !strings.HasPrefix(dark, "#") || len(dark) != 7 {
		t.Errorf("expected valid hex, got %s", dark)
	}
}

func TestColorForBranch_AllPaletteReachable(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		c := ColorForBranch(fmt.Sprintf("branch-%d", i))
		seen[c] = true
	}
	// With 16 colors and 1000 branches, we should see most of the palette
	if len(seen) < 10 {
		t.Errorf("poor distribution: only %d of 16 palette colors seen in 1000 branches", len(seen))
	}
}
