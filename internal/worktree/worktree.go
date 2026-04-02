package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// Create adds a git worktree at path. If newBranch is true, it creates a new
// branch from base. Otherwise it checks out an existing branch.
func Create(path, branch string, newBranch bool, base string) error {
	args := []string{"worktree", "add"}
	if newBranch {
		args = append(args, path, "-b", branch)
		if base != "" {
			args = append(args, base)
		}
	} else {
		args = append(args, path, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// BranchExists checks whether a branch exists locally or as a remote tracking ref.
func BranchExists(branch string) bool {
	if localBranchExists(branch) {
		return true
	}
	return remoteBranchExists(branch)
}

func localBranchExists(branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return cmd.Run() == nil
}

func remoteBranchExists(branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	return cmd.Run() == nil
}

// Fetch fetches a branch from the given remote.
func Fetch(remote, branch string) error {
	cmd := exec.Command("git", "fetch", remote, branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch %s %s failed: %s", remote, branch, strings.TrimSpace(string(out)))
	}
	return nil
}

// DetectMainRepo returns the root worktree path (the main repo) by parsing
// `git worktree list --porcelain`. Falls back to the given path.
func DetectMainRepo(worktreePath string) string {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = worktreePath
	out, err := cmd.Output()
	if err != nil {
		return worktreePath
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 && strings.HasPrefix(lines[0], "worktree ") {
		return strings.TrimPrefix(lines[0], "worktree ")
	}
	return worktreePath
}
