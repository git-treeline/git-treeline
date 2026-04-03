package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/setup"
	"github.com/git-treeline/git-treeline/internal/worktree"
	"github.com/spf13/cobra"
)

var newBase string
var newPath string
var newStart bool
var newDryRun bool

func init() {
	newCmd.Flags().StringVar(&newBase, "base", "", "Base branch for the new worktree (default: current branch)")
	newCmd.Flags().StringVar(&newPath, "path", "", "Custom worktree path (default: ../<project>-<branch>)")
	newCmd.Flags().BoolVar(&newStart, "start", false, "Run start_command after setup")
	newCmd.Flags().BoolVar(&newDryRun, "dry-run", false, "Print what would happen without making changes")
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a worktree, allocate resources, and run setup",
	Long: `Create a new git worktree for the given branch, allocate ports/databases/Redis,
and run setup commands. Combines 'git worktree add' with 'git-treeline setup' in one step.

If the branch already exists locally or on origin, it is checked out.
Otherwise a new branch is created from --base (or the current branch).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getting working directory: %w", err)
		}
		mainRepo := worktree.DetectMainRepo(cwd)
		pc := config.LoadProjectConfig(mainRepo)
		projectName := pc.Project()

		wtPath := newPath
		if wtPath == "" {
			wtPath = filepath.Join(filepath.Dir(mainRepo), fmt.Sprintf("%s-%s", projectName, branch))
		}

		// Check if this branch is already checked out in another worktree
		if existingWT := worktree.FindWorktreeForBranch(branch); existingWT != "" {
			return fmt.Errorf("branch '%s' is already checked out at %s\nUse 'git-treeline setup %s' to re-run setup on it", branch, existingWT, existingWT)
		}

		existing := worktree.BranchExists(branch)

		if newDryRun {
			if existing {
				fmt.Printf("[dry-run] Would check out existing branch '%s'\n", branch)
			} else {
				base := newBase
				if base == "" {
					base = "(current branch)"
				}
				fmt.Printf("[dry-run] Would create new branch '%s' from %s\n", branch, base)
			}
			fmt.Printf("[dry-run] Worktree path: %s\n", wtPath)
			fmt.Println("[dry-run] Would run: git-treeline setup")
			if newStart && pc.StartCommand() != "" {
				fmt.Printf("[dry-run] Would run: %s\n", pc.StartCommand())
			}
			return nil
		}

		if existing {
			if err := worktree.Fetch("origin", branch); err != nil {
				// Non-fatal: branch may only exist locally
				_ = err
			}
			fmt.Printf("==> Checking out existing branch '%s'\n", branch)
			if err := worktree.Create(wtPath, branch, false, ""); err != nil {
				return err
			}
		} else {
			base := newBase
			if base == "" {
				base = currentBranch()
			}
			fmt.Printf("==> Creating branch '%s' from '%s'\n", branch, base)
			if err := worktree.Create(wtPath, branch, true, base); err != nil {
				return err
			}
		}

		fmt.Printf("==> Worktree created at %s\n", wtPath)
		fmt.Println("==> Running setup...")

		uc := config.LoadUserConfig("")
		s := setup.New(wtPath, mainRepo, uc)
		s.Options.DryRun = false
		if _, err := s.Run(); err != nil {
			return fmt.Errorf("setup failed: %w", err)
		}

		if newStart {
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

func currentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "main"
	}
	return string(out[:len(out)-1]) // trim trailing newline
}

func execInWorktree(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
