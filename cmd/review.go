package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/github"
	"github.com/git-treeline/git-treeline/internal/setup"
	"github.com/git-treeline/git-treeline/internal/worktree"
	"github.com/spf13/cobra"
)

var reviewPath string
var reviewStart bool

func init() {
	reviewCmd.Flags().StringVar(&reviewPath, "path", "", "Custom worktree path (default: ../<project>-pr-<number>)")
	reviewCmd.Flags().BoolVar(&reviewStart, "start", false, "Run start_command after setup")
	rootCmd.AddCommand(reviewCmd)
}

var reviewCmd = &cobra.Command{
	Use:   "review <PR#>",
	Short: "Check out a pull request into a worktree and run setup",
	Long: `Fetch a GitHub pull request branch, create a worktree for it, allocate
resources, and run setup. Requires the gh CLI (https://cli.github.com).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PR number: %s", args[0])
		}

		fmt.Printf("==> Looking up PR #%d...\n", prNumber)
		pr, err := github.LookupPR(prNumber)
		if err != nil {
			return err
		}

		branch := pr.HeadRefName
		fmt.Printf("==> PR #%d → branch '%s'\n", prNumber, branch)

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		mainRepo := worktree.DetectMainRepo(cwd)
		pc := config.LoadProjectConfig(mainRepo)
		projectName := pc.Project()

		wtPath := reviewPath
		if wtPath == "" {
			wtPath = filepath.Join(filepath.Dir(mainRepo), fmt.Sprintf("%s-pr-%d", projectName, prNumber))
		}

		fmt.Printf("==> Fetching origin/%s...\n", branch)
		if err := worktree.Fetch("origin", branch); err != nil {
			return err
		}

		fmt.Printf("==> Creating worktree at %s\n", wtPath)
		if err := worktree.Create(wtPath, branch, false, ""); err != nil {
			return err
		}

		fmt.Println("==> Running setup...")
		uc := config.LoadUserConfig("")
		s := setup.New(wtPath, mainRepo, uc)
		alloc, err := s.Run()
		if err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		fmt.Println()
		fmt.Printf("PR #%d ready for review:\n", prNumber)
		fmt.Printf("  Branch:   %s\n", branch)
		fmt.Printf("  Path:     %s\n", wtPath)
		if alloc != nil {
			fmt.Printf("  URL:      http://localhost:%d\n", alloc.Port)
		}

		if reviewStart {
			startCmd := pc.StartCommand()
			if startCmd == "" {
				fmt.Println("Warning: --start passed but no start_command configured in .treeline.yml")
				return nil
			}
			fmt.Printf("==> Starting: %s\n", startCmd)
			return execInWorktree(wtPath, startCmd)
		}

		return nil
	},
}
