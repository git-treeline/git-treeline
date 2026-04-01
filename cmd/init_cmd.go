package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/git-treeline/git-treeline/internal/config"
	"github.com/git-treeline/git-treeline/internal/platform"
	"github.com/spf13/cobra"
)

var initProject string
var initTemplateDB string

func init() {
	initCmd.Flags().StringVar(&initProject, "project", "", "Project name")
	initCmd.Flags().StringVar(&initTemplateDB, "template-db", "", "Template database name for cloning")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a .treeline.yml config file for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := filepath.Join(".", config.ProjectConfigFile)
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintln(os.Stderr, ".treeline.yml already exists")
			os.Exit(1)
		}

		uc := config.LoadUserConfig("")
		if !uc.Exists() {
			if err := uc.Init(); err != nil {
				return err
			}
			fmt.Printf("==> Created user config at %s\n", platform.ConfigFile())
		}

		project := initProject
		if project == "" {
			cwd, _ := os.Getwd()
			project = filepath.Base(cwd)
		}

		templateDB := initTemplateDB
		if templateDB == "" {
			templateDB = project + "_development"
		}

		content := fmt.Sprintf(`project: %s

# Number of ports to allocate per worktree (e.g. 2 for app + esbuild reload)
# ports_needed: 1

# Environment file configuration
# target: file written in the worktree (e.g. .env.local, .env.development.local, .env)
# source: file copied from main repo as a starting point
env_file:
  target: .env.local
  source: .env.local

database:
  adapter: postgresql
  template: %s
  pattern: "{template}_{worktree}"

copy_files:
  - config/master.key

env:
  PORT: "{port}"
  DATABASE_NAME: "{database}"
  REDIS_URL: "{redis_url}"
  APPLICATION_HOST: "localhost:{port}"

setup_commands:
  - bundle install --quiet
  # - yarn install --silent

editor:
  vscode_title: '{project} (:{port}) — {branch} — ${activeEditorShort}'
`, project, templateDB)

		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}

		fmt.Printf("==> Created %s for project '%s'\n", config.ProjectConfigFile, project)
		fmt.Println()
		fmt.Printf("Allocation policy (ports, Redis) is managed in your user config:\n")
		fmt.Printf("  %s\n", platform.ConfigFile())

		openInEditor(path)
		return nil
	},
}

func openInEditor(path string) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return
	}
	exec.Command(editor, path).Run()
}
