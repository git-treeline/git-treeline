package github

import (
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
