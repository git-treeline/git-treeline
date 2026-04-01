package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/platform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or initialize user-level config",
	RunE: func(cmd *cobra.Command, args []string) error {
		uc := config.LoadUserConfig("")
		if !uc.Exists() {
			if err := uc.Init(); err != nil {
				return err
			}
			fmt.Printf("Created config at %s\n", platform.ConfigFile())
			return nil
		}

		fmt.Printf("Config: %s\n", platform.ConfigFile())
		data, _ := json.MarshalIndent(uc.Data, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}
