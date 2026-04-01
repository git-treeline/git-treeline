package cmd

import (
	"fmt"
	"os"

	"github.com/git-treeline/git-treeline/internal/registry"
	"github.com/spf13/cobra"
)

var pruneStale bool

func init() {
	pruneCmd.Flags().BoolVar(&pruneStale, "stale", false, "Also remove allocations for directories not listed in git worktree list")
	rootCmd.AddCommand(pruneCmd)
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove allocations for worktrees that no longer exist on disk",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.New("")

		var count int
		var err error
		if pruneStale {
			count, err = reg.PruneStale()
		} else {
			count, err = reg.Prune()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}

		if count == 0 {
			fmt.Println("Nothing to prune.")
		} else {
			fmt.Printf("Pruned %d stale allocation(s).\n", count)
		}
		return nil
	},
}
