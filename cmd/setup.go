package cmd

import (
	"fmt"
	"os"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/setup"
	"github.com/spf13/cobra"
)

var setupMainRepo string
var setupDryRun bool

func init() {
	setupCmd.Flags().StringVar(&setupMainRepo, "main-repo", "", "Path to the main repository (auto-detected if omitted)")
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false, "Print what would be allocated without writing anything")
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup [PATH]",
	Short: "Allocate resources and set up a worktree environment",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		uc := config.LoadUserConfig("")
		s := setup.New(path, setupMainRepo, uc)
		s.Options.DryRun = setupDryRun
		_, err := s.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return nil
	},
}
