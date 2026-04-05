package templates

import (
	"fmt"
	"strings"

	"github.com/git-treeline/git-treeline/internal/detect"
)

// TunnelHint returns framework-specific guidance for whitelisting a tunnel
// domain so the dev server accepts requests from it. Returns empty string
// if no configuration is needed.
//
// For named tunnels, pass the full hostname (e.g. "myapp-feature.example.dev")
// and domain (e.g. "example.dev"). For quick tunnels, pass hostname="" and
// domain="" to get a generic hint.
func TunnelHint(det *detect.Result, hostname, domain string) string {
	if hostname == "" {
		return tunnelHintQuick(det)
	}
	return tunnelHintNamed(det, hostname, domain)
}

func tunnelHintNamed(det *detect.Result, hostname, domain string) string {
	wildcard := "." + domain

	switch det.Framework {
	case "rails":
		return fmt.Sprintf(`Rails blocks requests from unknown hosts in development.
  Add to config/environments/development.rb:

    config.hosts << "%s"`, wildcard)

	case "vite":
		return fmt.Sprintf(`Vite's dev server rejects non-localhost hostnames (403).
  Add to your vite.config server options:

    server: { allowedHosts: ["%s"] }`, wildcard)

	case "django", "python":
		return fmt.Sprintf(`Django rejects requests when the Host header isn't in ALLOWED_HOSTS.
  Add to settings.py (or your dev settings):

    ALLOWED_HOSTS += ["%s"]`, hostname)

	default:
		return ""
	}
}

func tunnelHintQuick(det *detect.Result) string {
	switch det.Framework {
	case "rails":
		return `Rails blocks requests from unknown hosts in development.
  Quick tunnels get a random domain each time, so a wildcard is easiest:

    # config/environments/development.rb
    config.hosts << ".trycloudflare.com"

  For a stable domain, run 'gtl tunnel setup' instead.`

	case "vite":
		return `Vite's dev server rejects non-localhost hostnames (403).
  Quick tunnels get a random domain each time, so a wildcard is easiest:

    // vite.config
    server: { allowedHosts: [".trycloudflare.com"] }

  For a stable domain, run 'gtl tunnel setup' instead.`

	case "django", "python":
		return `Django rejects requests when the Host header isn't in ALLOWED_HOSTS.
  Quick tunnels get a random domain, so for dev you may need:

    ALLOWED_HOSTS += [".trycloudflare.com"]

  For a stable domain, run 'gtl tunnel setup' instead.`

	default:
		return ""
	}
}

// FormatTunnelHint wraps a TunnelHint in a visible block for terminal output.
// Returns empty string if there's no hint.
func FormatTunnelHint(hint string) string {
	if hint == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	for _, line := range strings.Split(hint, "\n") {
		b.WriteString("  " + line + "\n")
	}
	return b.String()
}
