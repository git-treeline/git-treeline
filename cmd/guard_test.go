package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	skipIfNoGit(t)
	dir := t.TempDir()
	dir, _ = filepath.EvalSymlinks(dir)
	runGit(t, dir, "init", "--initial-branch=main")
	runGit(t, dir, "commit", "--allow-empty", "-m", "init")
	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %s", args, string(out))
	}
}

func TestWorktreeGuard_AllowsMainRepo(t *testing.T) {
	repo := initRepo(t)

	orig, _ := os.Getwd()
	_ = os.Chdir(repo)
	defer func() { _ = os.Chdir(orig) }()

	err := worktreeGuard(newCmd, []string{"feature-x"})
	if err != nil {
		t.Errorf("expected no error from main repo, got: %v", err)
	}
}

func TestWorktreeGuard_BlocksWorktree(t *testing.T) {
	repo := initRepo(t)

	wtPath := filepath.Join(filepath.Dir(repo), "test-wt")
	runGit(t, repo, "worktree", "add", "-b", "test-branch", wtPath)
	defer func() {
		cmd := exec.Command("git", "worktree", "remove", "--force", wtPath)
		cmd.Dir = repo
		_ = cmd.Run()
	}()

	orig, _ := os.Getwd()
	_ = os.Chdir(wtPath)
	defer func() { _ = os.Chdir(orig) }()

	err := worktreeGuard(newCmd, []string{"feature-x"})
	if err == nil {
		t.Fatal("expected error from inside worktree")
	}
	if !strings.Contains(err.Error(), "not the main repo") {
		t.Errorf("expected 'not the main repo' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "gtl switch") {
		t.Errorf("expected 'gtl switch' suggestion in error, got: %v", err)
	}
}
