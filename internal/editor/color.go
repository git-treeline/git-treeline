// Package editor provides IDE integration for worktree environments.
// It generates per-worktree settings for VS Code/Cursor and JetBrains,
// including window titles, color theming, and workspace file detection.
package editor

import (
	"crypto/sha256"
	"fmt"
)

// Curated palette of visually distinct, accessible colors.
// Each has good contrast with white text and is distinguishable at a glance.
var palette = []string{
	"#1a5276", // deep blue
	"#7b241c", // crimson
	"#196f3d", // forest green
	"#7d3c98", // purple
	"#b9770e", // amber
	"#148f77", // teal
	"#a93226", // brick red
	"#1f618d", // steel blue
	"#6c3483", // violet
	"#117a65", // emerald
	"#b7950b", // gold
	"#2e4053", // dark slate
	"#884ea0", // orchid
	"#d4ac0d", // mustard
	"#2874a6", // royal blue
	"#c0392b", // scarlet
}

// ColorForBranch returns a deterministic hex color for the given branch name.
// The same branch always maps to the same color across sessions.
func ColorForBranch(branch string) string {
	if branch == "" {
		return palette[0]
	}
	h := sha256.Sum256([]byte(branch))
	idx := int(h[0]) % len(palette)
	return palette[idx]
}

// DarkenHex returns a slightly darker version of a hex color, for inactive states.
func DarkenHex(hex string) string {
	r, g, b := parseHex(hex)
	r = r * 85 / 100
	g = g * 85 / 100
	b = b * 85 / 100
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// ForegroundForHex returns white or dark text depending on background luminance.
func ForegroundForHex(hex string) string {
	r, g, b := parseHex(hex)
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	if luminance > 160 {
		return "#1e1e1e"
	}
	return "#ffffff"
}

func parseHex(hex string) (int, int, int) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 0, 0, 0
	}
	var r, g, b int
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
