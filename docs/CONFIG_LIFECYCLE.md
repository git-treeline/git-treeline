# How `.treeline.yml` Changes Take Effect

This document explains when edits to `.treeline.yml` are applied, which commands trigger them, and what requires manual action.

## The Rule

**Each worktree has its own `.treeline.yml`** (it's a checked-out file from that branch). Commands always read from the worktree you're in — not the main repo.

Edit `.treeline.yml` in your worktree, then use `gtl start`, `gtl env sync`, or the appropriate command to apply changes.

---

## Quick Reference

| Field | Applied on fresh `gtl start`? | How to apply manually |
|-------|-------------------------------|----------------------|
| `env:` | Yes (also on `gtl restart`) | `gtl env sync` |
| `commands.start` | Yes (fresh start only — see note) | Ctrl+C supervisor, then `gtl start` |
| `editor` | Yes (also via `gtl env sync`) | `gtl editor refresh` |
| `commands.setup` | No — slow/destructive | `gtl setup` or `gtl switch --setup` |
| `copy_files` | No — provisioning only | `gtl setup` |
| `database.*` | No — destructive | `gtl db reset` |
| `port_count` | No — registry conflict risk | `gtl setup` |
| `project` | Yes (read on demand) | Next command that uses it |
| `hooks` | N/A | Fires when the hook event occurs |

**Note on `commands.start`:** `gtl start` reads the new command only when no supervisor is running. If the supervisor is already alive (even if the server was stopped via `gtl stop`), `gtl start` resumes with the original command — `gtl stop` stops the child process but the supervisor stays alive. To pick up a new `commands.start`, Ctrl+C the supervisor terminal and run `gtl start` fresh. (`gtl restart` will warn you if it detects a mismatch.)

---

## Scenarios

### "I changed an env var"

You edited the `env:` block in `.treeline.yml` — for example, changed `APPLICATION_HOST`.

- **`gtl start`** (fresh — no supervisor running) syncs the env file (e.g. `.env.local`), then starts the server. Your change is live. If the supervisor is already alive (e.g. after `gtl stop`), `gtl start` resumes without re-syncing — use `gtl restart` instead.
- **`gtl restart`** syncs the env file and updates the running supervisor's environment, then restarts. Your change is live.
- **`gtl env sync`** syncs the env file without touching the server. For users who start their server outside of `gtl`.

### "I changed `commands.start`"

You changed the start command — for example, from `rails server` to `bin/dev`.

- **Fresh `gtl start`** (no supervisor running) reads the new command and uses it.
- **`gtl start`** (supervisor still alive, e.g. after `gtl stop`) resumes with the original command — the supervisor holds it in memory. `gtl stop` only kills the child process; the supervisor stays running.
- **`gtl restart`** does NOT pick up the new command either, but will warn you about the mismatch.
- **To apply:** Ctrl+C the supervisor terminal, then run `gtl start` again.

### "I changed `editor.title` or `editor.color`"

- **`gtl start`** (fresh — no supervisor running) syncs editor settings before starting.
- **`gtl editor refresh`** updates editor settings without starting a server.
- **`gtl env sync`** also refreshes editor settings.

### "I changed the database config"

You changed `database.adapter`, `database.template`, or `database.pattern`.

- **Nothing happens automatically.** Database operations are destructive.
- Databases are cloned once during `gtl new` (worktree creation).
- To drop and re-clone from the template: **`gtl db reset`**.
- To re-clone from a different source: **`gtl db reset --from other_db`**.

### "I changed `copy_files`"

You added a new file to the `copy_files` list.

- **`gtl setup`** copies files from the main repo into the worktree.
- Not automatic on `gtl start` — `copy_files` is provisioning, not runtime.
- The file list comes from your worktree's config. The actual file content comes from the main repo (these are typically shared secrets like `config/master.key`).

### "I changed `commands.setup`"

You updated setup commands — for example, added `yarn install`.

- **`gtl setup`** runs the new commands.
- **`gtl switch --setup`** runs setup after switching branches.
- Not automatic on `gtl start` — setup commands can be slow (bundle install, migrations).

### "I changed `port_count`"

You need an extra port — for example, for esbuild.

- **`gtl setup`** detects the mismatch and re-allocates ports.
- Not automatic on `gtl start` — port changes affect the shared allocation registry and could conflict with other worktrees.

---

## Commands and What They Read

| Command | Config source | What it does with `.treeline.yml` |
|---------|--------------|----------------------------------|
| `gtl start` | Worktree | Fresh: syncs env file, syncs editor, reads `commands.start`. Resume: sends start signal only |
| `gtl restart` | Worktree | Syncs env file, updates supervisor env, restarts server |
| `gtl stop` | N/A | Sends stop signal via socket — no config read |
| `gtl env sync` | Worktree | Syncs env file and editor settings |
| `gtl env show` | Worktree | Reads `env:` and `env_file` to display current state |
| `gtl setup` | Worktree | Full provisioning: allocate, copy files, write env, clone DB, run setup commands |
| `gtl new` | Main repo → Worktree | Reads main repo for allocation, then runs setup from the new worktree |
| `gtl review` | Main repo → Worktree | Same as `gtl new` |
| `gtl switch` | Worktree | Refreshes env, optionally re-runs setup |
| `gtl refresh` | Worktree | Detects port changes using worktree's `port_count` |
| `gtl open` | Worktree | Reads `project` for URL construction |
| `gtl tunnel` | Worktree | Reads `project` for subdomain routing |
| `gtl db` | Worktree | Reads `database.*` for adapter, template, target |
| `gtl release` | Worktree | Reads `hooks` for pre/post_release |
| `gtl doctor` | Worktree | Reads all fields for diagnostics |
| `gtl editor refresh` | Worktree | Reads `editor.*` |
| `gtl init` | Main repo | Writes new config to main repo |
| `gtl prune` | Worktree (partial) | Registry operation; reads `merge_target` from `.treeline.yml` for `--merged` |

---

**Domain migration:** New installs default to `prt.dev` instead of `localhost`. See [DOMAIN_MIGRATION.md](DOMAIN_MIGRATION.md) for details, tradeoffs, and upgrade steps. Existing installs are not affected — the CLI detects pre-existing configs and preserves `localhost`.
