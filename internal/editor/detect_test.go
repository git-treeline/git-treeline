package editor

import "testing"

func TestLookupEditor_Known(t *testing.T) {
	cases := []struct {
		name    string
		display string
		cli     string
	}{
		{"cursor", "Cursor", "cursor"},
		{"vscode", "VS Code", "code"},
		{"zed", "Zed", "zed"},
		{"goland", "GoLand", "goland"},
		{"webstorm", "WebStorm", "webstorm"},
		{"rubymine", "RubyMine", "rubymine"},
		{"idea", "IntelliJ IDEA", "idea"},
		{"sublime", "Sublime Text", "subl"},
		{"neovim", "Neovim", "nvim"},
	}

	for _, tc := range cases {
		info := LookupEditor(tc.name)
		if info == nil {
			t.Errorf("LookupEditor(%q) returned nil", tc.name)
			continue
		}
		if info.Display != tc.display {
			t.Errorf("LookupEditor(%q).Display = %q, want %q", tc.name, info.Display, tc.display)
		}
		if info.CLI != tc.cli {
			t.Errorf("LookupEditor(%q).CLI = %q, want %q", tc.name, info.CLI, tc.cli)
		}
	}
}

func TestLookupEditor_Unknown(t *testing.T) {
	if info := LookupEditor("notepad++"); info != nil {
		t.Errorf("expected nil for unknown editor, got %+v", info)
	}
	if info := LookupEditor(""); info != nil {
		t.Errorf("expected nil for empty name, got %+v", info)
	}
}

func TestJetbrainsProduct(t *testing.T) {
	cases := map[string]string{
		"GoLand":        "goland",
		"WebStorm":      "webstorm",
		"RubyMine":      "rubymine",
		"IntelliJIdea":  "idea",
		"PyCharm":       "pycharm",
		"PhpStorm":      "phpstorm",
		"UnknownFuture": "idea",
	}
	for input, want := range cases {
		got := jetbrainsProduct(input)
		if got != want {
			t.Errorf("jetbrainsProduct(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCliToName(t *testing.T) {
	cases := map[string]string{
		"code":    "vscode",
		"subl":    "sublime",
		"nvim":    "neovim",
		"cursor":  "cursor",
		"zed":     "zed",
		"goland":  "goland",
		"unknown": "",
	}
	for cli, want := range cases {
		got := cliToName(cli)
		if got != want {
			t.Errorf("cliToName(%q) = %q, want %q", cli, got, want)
		}
	}
}

func TestDetectEditor_ReturnsString(t *testing.T) {
	// DetectEditor should always return a string (possibly empty).
	// We can't control env vars in a unit test without side effects,
	// but we can verify it doesn't panic.
	result := DetectEditor()
	_ = result
}
