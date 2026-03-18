package app

import (
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ia <agent> <language> <project> [--dry-run] [--mask-file <path>] [--mask-dir <path>]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  --dry-run              print docker command without running it")
	fmt.Fprintln(os.Stderr, "  --mask-file <path>     mount target file as /dev/null inside the container")
	fmt.Fprintln(os.Stderr, "  --mask-dir <path>      mount target directory as tmpfs inside the container")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Agents:   %s\n", agentList())
}

type runOptions struct {
	dryRun    bool
	nullFiles MountTargets
	tmpfsDirs MountTargets
}

func Run(args []string) int {
	if len(args) < 4 {
		usage()
		return 2
	}

	agentName := args[1]
	language := args[2]
	project := args[3]

	opts, err := parseRunOptions(args[4:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		usage()
		return 2
	}

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

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		return 1
	}

	cfg.applyRunOptions(opts)

	if err := cfg.validateForAgent(agentName); err != nil {
		fmt.Fprintf(os.Stderr, "error: validate config: %v\n", err)
		return 1
	}

	argsToRun := buildArgs(cfg, agentName, language, project)
	return run(argsToRun, opts.dryRun)
}

func parseRunOptions(args []string) (runOptions, error) {
	var opts runOptions

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--dry-run":
			opts.dryRun = true
		case arg == "--mask-file":
			if i+1 >= len(args) {
				return runOptions{}, fmt.Errorf("option %q requires a value", arg)
			}
			i++
			opts.nullFiles.Merge(parseMountTargets(args[i]))
		case strings.HasPrefix(arg, "--mask-file="):
			opts.nullFiles.Merge(parseMountTargets(strings.TrimPrefix(arg, "--mask-file=")))
		case arg == "--mask-dir":
			if i+1 >= len(args) {
				return runOptions{}, fmt.Errorf("option %q requires a value", arg)
			}
			i++
			opts.tmpfsDirs.Merge(parseMountTargets(args[i]))
		case strings.HasPrefix(arg, "--mask-dir="):
			opts.tmpfsDirs.Merge(parseMountTargets(strings.TrimPrefix(arg, "--mask-dir=")))
		default:
			return runOptions{}, fmt.Errorf("unknown option %q", arg)
		}
	}

	return opts, nil
}
