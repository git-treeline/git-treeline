package editor

import (
	"os"
	"path/filepath"
	"strings"
)

// WriteJetBrains sets the project header color in .idea/workspace.xml.
// JetBrains IDEs (2023.2+) support a colored header bar per project.
func WriteJetBrains(worktreePath, color string) (string, error) {
	ideaDir := filepath.Join(worktreePath, ".idea")
	wsPath := filepath.Join(ideaDir, "workspace.xml")

	if _, err := os.Stat(ideaDir); os.IsNotExist(err) {
		return "", nil
	}

	content, err := os.ReadFile(wsPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	updated := setProjectHeaderColor(string(content), color)
	if err := os.WriteFile(wsPath, []byte(updated), 0o644); err != nil {
		return "", err
	}
	return wsPath, nil
}

// DetectJetBrains returns true if the worktree has a .idea directory.
func DetectJetBrains(worktreePath string) bool {
	_, err := os.Stat(filepath.Join(worktreePath, ".idea"))
	return err == nil
}

// setProjectHeaderColor uses targeted string manipulation instead of full XML
// round-tripping to avoid destroying non-component elements in workspace.xml.
func setProjectHeaderColor(xmlContent, color string) string {
	if xmlContent == "" {
		return buildMinimalWorkspaceXML(color)
	}

	colorVal := strings.TrimPrefix(color, "#")
	optionTag := `<option name="customColor" value="` + colorVal + `" />`
	componentBlock := "  <component name=\"ProjectColorInfo\">\n    " + optionTag + "\n  </component>"

	if strings.Contains(xmlContent, `name="ProjectColorInfo"`) {
		// Replace existing customColor value
		lines := strings.Split(xmlContent, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, `name="customColor"`) &&
				strings.Contains(trimmed, `<option`) {
				indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
				lines[i] = indent + optionTag
				return strings.Join(lines, "\n")
			}
		}
		// Component exists but no customColor option — insert one
		for i, line := range lines {
			if strings.Contains(line, `name="ProjectColorInfo"`) {
				indent := "    "
				insert := indent + optionTag
				rest := make([]string, len(lines)-i-1)
				copy(rest, lines[i+1:])
				lines = append(lines[:i+1], insert)
				lines = append(lines, rest...)
				return strings.Join(lines, "\n")
			}
		}
	}

	// No ProjectColorInfo component — insert before closing </project>
	closing := "</project>"
	idx := strings.LastIndex(xmlContent, closing)
	if idx < 0 {
		return buildMinimalWorkspaceXML(color)
	}
	return xmlContent[:idx] + componentBlock + "\n" + closing + "\n"
}

func buildMinimalWorkspaceXML(color string) string {
	colorVal := strings.TrimPrefix(color, "#")
	return `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectColorInfo">
    <option name="customColor" value="` + colorVal + `" />
  </component>
</project>
`
}
