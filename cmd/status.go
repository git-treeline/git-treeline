package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/git-treeline/git-treeline/internal/registry"
	"github.com/spf13/cobra"
)

var statusProject string
var statusJSON bool

func init() {
	statusCmd.Flags().StringVar(&statusProject, "project", "", "Filter by project name")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all active allocations across projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := registry.New("")
		allocs := reg.Allocations()
		if statusProject != "" {
			allocs = reg.FindByProject(statusProject)
		}

		if statusJSON {
			data, _ := json.MarshalIndent(allocs, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(allocs) == 0 {
			fmt.Println("No active allocations.")
			return nil
		}

		grouped := make(map[string][]registry.Allocation)
		for _, a := range allocs {
			project := ""
			if p, ok := a["project"].(string); ok {
				project = p
			}
			grouped[project] = append(grouped[project], a)
		}

		for project, entries := range grouped {
			sort.Slice(entries, func(i, j int) bool {
				pi, _ := entries[i]["port"].(float64)
				pj, _ := entries[j]["port"].(float64)
				return pi < pj
			})

			fmt.Printf("\n%s:\n", project)
			for _, a := range entries {
				ports := getPorts(a)
				portLabel := joinInts(ports, ",")

				name, _ := a["worktree_name"].(string)
				db, _ := a["database"].(string)

				redis := ""
				if prefix, ok := a["redis_prefix"].(string); ok && prefix != "" {
					redis = "prefix:" + prefix
				} else if rdb, ok := a["redis_db"].(float64); ok {
					redis = fmt.Sprintf("db:%d", int(rdb))
				}

				fmt.Printf("  :%s  %s  db:%s  %s\n", portLabel, name, db, redis)
			}
		}

		return nil
	},
}
