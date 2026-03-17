package app

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  agent <agent> <language> <project> [--dry-run]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Agents:   %s\n", agentList())
}

func Run(args []string) int {
	if len(args) < 4 {
		usage()
		return 2
	}

	agentName := args[1]
	language := args[2]
	project := args[3]

	dryRun := len(args) > 4 && args[4] == "--dry-run"

	if !isSupportedAgent(agentName) {
		fmt.Fprintf(os.Stderr, "error: unknown agent %q\n", agentName)
		return 1
	}

	if !isValidLanguage(language) {
		fmt.Fprintf(os.Stderr, "error: invalid language %q\n", language)
		return 1
	}

	if !isValidProject(project) {
		fmt.Fprintf(os.Stderr, "error: invalid project %q\n", project)
		return 1
	}

	cfg, err := loadConfig(agentName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		return 1
	}

	argsToRun := buildArgs(cfg, agentName, language, project)
	return run(argsToRun, dryRun)
}
