package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/git-treeline/git-treeline/internal/format"
	"github.com/git-treeline/git-treeline/internal/registry"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(portCmd)
}

var portCmd = &cobra.Command{
	Use:   "port",
	Short: "Print the allocated port for the current worktree",
	Long:  `Prints the primary allocated port for the current directory's worktree. Useful for scripts, agents, and CI that need the port without parsing status output.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		absPath, _ := filepath.Abs(cwd)

		reg := registry.New("")
		entry := reg.Find(absPath)
		if entry == nil {
			fmt.Fprintf(os.Stderr, "No allocation found for %s\nRun `gtl setup` first.\n", absPath)
			os.Exit(1)
		}

		ports := format.GetPorts(format.Allocation(entry))
		if len(ports) == 0 {
			fmt.Fprintln(os.Stderr, "Allocation exists but has no ports.")
			os.Exit(1)
		}

		fmt.Println(ports[0])
		return nil
	},
}
