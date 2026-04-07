# Project Config Loading: Proposed Design

## Principle

**Always load `.treeline.yml` from the worktree you're operating on.**

The only exceptions are operations that run *before* a worktree exists or need files from the main repo.

---

## Command-by-Command Specification

### Commands That Should Load From Worktree

| Command | Current | Proposed | Fields Used |
|---------|---------|----------|-------------|
| `gtl start` | ✅ Worktree | ✅ Worktree | `commands.start`, `env` |
| `gtl stop/restart` | N/A | N/A | None (socket IPC only) |
| `gtl setup` | ✅ Worktree | ✅ Worktree | All fields |
| `gtl refresh` | ⚠️ Mixed | ✅ Worktree | `env`, `editor`, `port_count` |
| `gtl switch` | ✅ Worktree | ✅ Worktree | `env`, `editor`, optionally `commands.setup` |
| `gtl doctor` | ✅ Worktree | ✅ Worktree | All fields (read-only) |
| `gtl env` | ✅ Worktree | ✅ Worktree | `env`, `env_file` |
| `gtl editor refresh` | ✅ Worktree | ✅ Worktree | `editor` |
| `gtl link/unlink` | ✅ Worktree | ✅ Worktree | `env` |
| `gtl release` | ❌ Main repo | ✅ Worktree | `hooks` |
| `gtl open` | ❌ Main repo | ✅ Worktree | `project` |
| `gtl tunnel` | ❌ Main repo | ✅ Worktree | `project` |
| `gtl db` | ❌ Main repo | ✅ Worktree | `database.*` |

### Commands That Should Load From Main Repo

| Command | Reason |
|---------|--------|
| `gtl new` | Worktree doesn't exist yet; load main repo, then after `git worktree add`, run setup from the *new* worktree |
| `gtl review` | Same as `new` |
| `gtl clone` | Operating on the cloned repo itself (main = worktree) |
| `gtl init` | Writing config to main repo |
| `gtl prune` | Global operation, no specific worktree |

### Special Case: `copy_files`

During `gtl setup`, `copy_files` copies from **main repo** into the worktree. This is correct — the source files live in main repo. The *list* of files to copy should come from the worktree's `.treeline.yml`.

---

## When Changes to `.treeline.yml` Take Effect

| Field | Takes Effect On |
|-------|-----------------|
| `project` | Next command that reads it |
| `port_count` | `gtl setup` (re-allocates ports) |
| `database.*` | `gtl db reset` (explicit) or new worktree |
| `copy_files` | `gtl setup` (re-copies) |
| `env_file` | `gtl env sync` or `gtl start` |
| `env:` | `gtl env sync` or `gtl start` |
| `commands.setup` | `gtl setup` or `gtl switch --setup` |
| `commands.start` | Next `gtl start` |
| `editor` | `gtl editor refresh` or `gtl start` |
| `hooks` | When the hook event fires |

---

## New Behavior: `gtl start` Syncs Env

**Before starting the server, `gtl start` will:**

1. Re-read `env:` from worktree's `.treeline.yml`
2. Update the env file (e.g., `.env.local`) with current values
3. Update editor config
4. Then start the server

This ensures changes to `.treeline.yml` take effect without manual intervention.

**`gtl restart`** will also sync before restarting.

---

## New Command: `gtl env sync`

For users who don't use `gtl start`:

```
gtl env sync
```

Re-reads `env:` from `.treeline.yml` and updates the env file. No server interaction.

This replaces the confusing `gtl refresh` behavior (which only runs on port conflicts).

---

## Fixes Required

### High Priority (Broken Behavior)

1. **`gtl open`** — Change from `LoadProjectConfig(mainRepo)` to `LoadProjectConfig(absPath)`
2. **`gtl tunnel`** — Same fix
3. **`gtl db`** — Same fix
4. **`gtl release`** — Same fix (for hooks)
5. **`gtl refresh` detection** — Use worktree's `port_count`, not main repo's
6. **`gtl new --start`** — After worktree created, reload config from worktree for `StartCommand()`
7. **`gtl review --start`** — Same fix

### Medium Priority (New Features)

8. **`gtl start`** — Add env sync before starting server
9. **`gtl restart`** — Add env sync before restarting
10. **`gtl env sync`** — New command

### Low Priority (Cleanup)

11. **`gtl setup` diagnostics** — Load from worktree, not main repo
12. **Documentation** — Update all docs to reflect this model

---

## Migration

This is **not a breaking change** for most users. The behavior becomes more intuitive:
- Edit `.treeline.yml` in your worktree
- Run `gtl start` (or `gtl env sync`)
- Changes take effect

Users who relied on main repo config being used everywhere will see different behavior, but that was arguably a bug, not a feature.

---

## Summary

**One rule:** Load config from the worktree you're in.

**Two exceptions:** `gtl new`/`review` (worktree doesn't exist yet), `gtl init`/`prune` (global operations).

**One new behavior:** `gtl start` syncs env before starting.

**One new command:** `gtl env sync` for non-`gtl start` users.
