package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestCliErr_SuppressesUsageOnError(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliErr(cmd, &CliError{
				Message: "Something went wrong",
				Hint:    "Try again",
			})
		},
	}

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}

	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage to be true after cliErr")
	}
}

func TestCliErr_DoesNotSuppressOnNil(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cliErr(cmd, nil)
		},
	}

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd.SilenceUsage {
		t.Error("expected SilenceUsage to remain false when error is nil")
	}
}

func TestCliErr_PreservesUsageForCobraErrors(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	cmd.SetArgs([]string{}) // wrong number of args
	_ = cmd.Execute()

	// Cobra's arg validation error should NOT silence usage
	if cmd.SilenceUsage {
		t.Error("expected SilenceUsage to remain false for Cobra validation errors")
	}
}
