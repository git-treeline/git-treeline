---
description: Versioning, branching, and release conventions for git-treeline
globs: ["**"]
---

# Versioning

Semver (pre-1.0): minor bump for new features/commands, patch for bugfixes.

## Workflow

1. Branch from `main`: `feature/<name>` or `fix/<name>`
2. Implement, ensure `make ci` passes
3. Update `CHANGELOG.md` with new version heading and entries
4. PR, merge to `main`
5. Tag `vX.Y.Z` on main — triggers goreleaser via `.github/workflows/release.yml`
6. GoReleaser builds binaries and pushes Homebrew formula to `git-treeline/homebrew-tap`

## Rules

- Never tag without `make ci` passing
- Never skip CHANGELOG update
- One logical feature set per minor version
- Version string injected via ldflags: `-X github.com/git-treeline/git-treeline/cmd.Version={{ .Version }}`
