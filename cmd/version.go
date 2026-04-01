package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.3.0"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("git-treeline %s\n", Version)
	},
}
