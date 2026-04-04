package github

import (
	"os"
	"os/exec"
	"testing"
)

func TestCheckGH(t *testing.T) {
	_, err := exec.LookPath("gh")
	if err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	if err := checkGH(); err != nil {
		t.Errorf("checkGH failed when gh is installed: %v", err)
	}
}

func TestLookupPR_NotARepo(t *testing.T) {
	_, err := exec.LookPath("gh")
	if err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	// Run from a temp dir that isn't a git repo
	_, err = LookupPR(999999)
	if err == nil {
		t.Error("expected error for non-existent PR")
	}
}

func TestListOpenPRs_GracefulFailure(t *testing.T) {
	_, err := exec.LookPath("gh")
	if err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	// From a temp dir (not a git repo), should return nil without error
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	defer func() { _ = os.Chdir(orig) }()

	prs, err := ListOpenPRs()
	if err != nil {
		t.Errorf("expected nil error on graceful failure, got: %v", err)
	}
	// prs may be nil or empty — either is fine, just not a panic
	_ = prs
}

func TestListOpenPRs_ReturnsData(t *testing.T) {
	_, err := exec.LookPath("gh")
	if err != nil {
		t.Skip("gh CLI not installed, skipping")
	}

	// From the real repo, should succeed (may return empty if no open PRs)
	prs, err := ListOpenPRs()
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	// Just verify the call didn't panic and returned a valid type
	for _, pr := range prs {
		if pr.HeadRefName == "" {
			t.Error("expected non-empty HeadRefName for each PR")
		}
	}
}
