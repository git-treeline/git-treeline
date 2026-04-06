package tunnel

import (
	"os"
	"os/exec"
	"path/filepath"
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
		got := ExtractTrycloudflareURL(tc.line)
		if got != tc.want {
			t.Errorf("ExtractTrycloudflareURL(%q) = %q, want %q", tc.line, got, tc.want)
		}
	}
}

func TestFindCredentialsFile_NoFallbackScan(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := findCredentialsFile("my-tunnel")
	want := dir + "/.cloudflared/my-tunnel.json"
	if got != want {
		t.Errorf("findCredentialsFile = %q, want %q", got, want)
	}
}

func TestFilterLine_Errors(t *testing.T) {
	cases := []struct {
		line    string
		printed bool
	}{
		{"2024 ERR failed to connect", true},
		{"2024 WRN retrying in 5s", true},
		{"2024 INF Registered tunnel connection", true},
		{"2024 INF Starting tunnel", false},
		{"GET /api/health 200 12ms", true},
		{"POST /webhook 201 5ms", true},
		{"some other log line", false},
		{"connection failed to establish", true},
		{"error: dial tcp", true},
	}
	for _, tc := range cases {
		FilterLine(tc.line)
	}
}

func TestWriteTunnelConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	_ = os.MkdirAll(filepath.Join(dir, ".cloudflared"), 0o700)
	credPath := filepath.Join(dir, ".cloudflared", "test-tunnel.json")
	_ = os.WriteFile(credPath, []byte(`{"AccountTag":"abc"}`), 0o600)

	path, err := writeTunnelConfig("test-tunnel", "myapp-main.example.dev", 3050)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	checks := []string{
		`tunnel: "test-tunnel"`,
		`hostname: "myapp-main.example.dev"`,
		"http://localhost:3050",
		"http_status:404",
	}
	for _, check := range checks {
		if !strings.Contains(s, check) {
			t.Errorf("config missing %q\nGot:\n%s", check, s)
		}
	}
}

func TestConfigDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	got := ConfigDir()
	if !strings.HasSuffix(got, ".cloudflared") {
		t.Errorf("expected path ending in .cloudflared, got %s", got)
	}
}

func TestParseCertZoneID(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")

	// Valid cert with zone ID
	validCert := `-----BEGIN ARGO TUNNEL TOKEN-----
eyJ6b25lSUQiOiJhYmMxMjMiLCJhY2NvdW50SUQiOiJkZWY0NTYiLCJhcGlUb2tlbiI6InRlc3QifQ==
-----END ARGO TUNNEL TOKEN-----
`
	if err := os.WriteFile(certPath, []byte(validCert), 0o600); err != nil {
		t.Fatal(err)
	}

	zoneID, err := ParseCertZoneID(certPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if zoneID != "abc123" {
		t.Errorf("ParseCertZoneID = %q, want %q", zoneID, "abc123")
	}
}

func TestParseCertZoneID_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "cert.pem")

	if err := os.WriteFile(certPath, []byte("not a valid cert"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := ParseCertZoneID(certPath)
	if err == nil {
		t.Error("expected error for invalid cert format")
	}
}

func TestCertPathForDomain(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := CertPathForDomain("example.com")
	if !strings.HasSuffix(path, "cert-example.com.pem") {
		t.Errorf("CertPathForDomain = %q, expected suffix cert-example.com.pem", path)
	}
}

func TestIsLoggedInForDomain(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// No cert exists
	if IsLoggedInForDomain("example.com") {
		t.Error("expected false when no domain cert exists")
	}

	// Create domain cert
	cfDir := filepath.Join(dir, ".cloudflared")
	_ = os.MkdirAll(cfDir, 0o700)
	certPath := filepath.Join(cfDir, "cert-example.com.pem")
	_ = os.WriteFile(certPath, []byte("cert"), 0o600)

	if !IsLoggedInForDomain("example.com") {
		t.Error("expected true when domain cert exists")
	}
}
