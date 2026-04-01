package cmd

import (
	"fmt"
	"os"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/setup"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(refreshCmd)
}

var refreshCmd = &cobra.Command{
	Use:   "refresh [PATH]",
	Short: "Re-interpolate env file and editor config using existing allocation",
	Long: `Refresh re-reads .treeline.yml and rewrites the env file using the existing
port/database/Redis allocation. No new allocation is made, no database is
cloned, and no setup commands are re-run. Use this after updating .treeline.yml
to propagate changes to existing worktrees.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		uc := config.LoadUserConfig("")
		s := setup.New(path, "", uc)
		s.Options.RefreshOnly = true
		_, err := s.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return nil
	},
}
