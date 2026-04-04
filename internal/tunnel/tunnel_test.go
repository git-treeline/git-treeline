package tunnel

import (
	"os/exec"
	"strings"
	"testing"
)

func TestResolveCloudflared_ErrorWhenMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := ResolveCloudflared()
	if err == nil {
		t.Fatal("expected error when cloudflared is not in PATH")
	}
	if !strings.Contains(err.Error(), "not found in PATH") {
		t.Errorf("expected install instructions in error, got: %v", err)
	}
}

func TestResolveCloudflared_FindsIfPresent(t *testing.T) {
	if _, err := exec.LookPath("cloudflared"); err != nil {
		t.Skip("cloudflared not installed, skipping")
	}
	path, err := ResolveCloudflared()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestGenerateConfig(t *testing.T) {
	config := GenerateConfig("gtl", "salt-main.myteam.dev", 3050, "/home/user/.cloudflared/abc123.json")

	checks := []string{
		`tunnel: "gtl"`,
		`credentials-file: "/home/user/.cloudflared/abc123.json"`,
		`hostname: "salt-main.myteam.dev"`,
		"service: http://localhost:3050",
		"service: http_status:404",
	}
	for _, check := range checks {
		if !strings.Contains(config, check) {
			t.Errorf("config missing %q\nGot:\n%s", check, config)
		}
	}
}

func TestIsLoggedIn_FalseByDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if IsLoggedIn() {
		t.Error("expected IsLoggedIn to be false with empty home dir")
	}
}

func TestExtractTrycloudflareURL(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"2024-01-01 INF +----------------------------+", ""},
		{"2024-01-01 INF |  https://foo-bar-baz.trycloudflare.com  |", "https://foo-bar-baz.trycloudflare.com"},
		{"some random log line", ""},
		{"https://abc-123.trycloudflare.com is ready", "https://abc-123.trycloudflare.com"},
	}
	for _, tc := range cases {
		got := extractTrycloudflareURL(tc.line)
		if got != tc.want {
			t.Errorf("extractTrycloudflareURL(%q) = %q, want %q", tc.line, got, tc.want)
		}
	}
}

func TestFindCredentialsFile_NoFallbackScan(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Without a valid tunnel ID lookup, it should return a deterministic path,
	// not scan for random .json files.
	got := findCredentialsFile("my-tunnel")
	want := dir + "/.cloudflared/my-tunnel.json"
	if got != want {
		t.Errorf("findCredentialsFile = %q, want %q", got, want)
	}
}
