# Contributing to Git Treeline

Thanks for considering a contribution. Here's what you need to know.

## Setup

```bash
git clone https://github.com/git-treeline/git-treeline.git
cd git-treeline
make ci
```

Requires Go 1.24+ and optionally `golangci-lint` and `govulncheck` (auto-installed by `make ci` if missing).

## Making changes

1. Fork the repo and create a branch from `main`.
2. Write tests for new behavior.
3. Run `make ci` and ensure all checks pass.
4. Open a pull request with a clear description of the change and why it's needed.

## Pull request expectations

- One logical change per PR.
- Include a test plan in the PR description.
- Keep commits focused — squash fixups before requesting review.

## Reporting bugs

Open an issue with steps to reproduce, expected behavior, and actual behavior. Include your OS and `git-treeline version` output.

## Security vulnerabilities

Please report security issues privately — see [SECURITY.md](SECURITY.md).
