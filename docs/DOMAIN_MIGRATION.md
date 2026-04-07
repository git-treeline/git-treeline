# Domain Migration: `localhost` → `prt.dev`

New installs default to `prt.dev` as the router domain. Existing installs are not affected.

## Why the change

Chrome and Brave strip `https://` from `.localhost` URLs when opened via the macOS `open` command, terminal links, or paste. The URL resolves correctly if you type it manually, but clickable links (from `gtl open`, `gtl start` output, or IDE terminals) open as `http://` and fail. This is a browser-level quirk specific to the `.localhost` TLD.

`prt.dev` is a domain with wildcard DNS (`*.prt.dev → 127.0.0.1`). Browsers treat it as a normal HTTPS domain, so links work correctly.

## Tradeoff: DNS dependency

`.localhost` resolves locally with zero external dependency. `prt.dev` depends on external DNS remaining pointed at `127.0.0.1`. If DNS is unreachable (airplane mode with no cache, domain lapse), local HTTPS URLs won't resolve.

For users who prefer zero external dependency and can tolerate the Chrome link issue, `localhost` remains fully supported:

```bash
gtl config set router.domain localhost
```

## What happens to existing installs

**Nothing breaks.** The CLI detects whether your `config.json` already exists:

- **Config exists, `router.domain` absent** → defaults to `localhost` (preserves pre-upgrade behavior)
- **Config exists, `router.domain` set** → uses whatever you set
- **No config at all** (fresh machine) → defaults to `prt.dev`

Additionally, `gtl serve install` now persists `router.domain` to `config.json` at install time. After installing, the domain is always explicit on disk — no reliance on code defaults.

## If you want to switch to `prt.dev`

```bash
gtl serve uninstall
gtl config set router.domain prt.dev
gtl serve install
```

This regenerates certificates for the new domain and updates port forwarding. If you have OAuth redirect URIs registered to `*.localhost`, update those to `*.prt.dev`.

## If you want to stay on `localhost`

No action needed — existing installs already default to `localhost`. To make it explicit:

```bash
gtl config set router.domain localhost
```
