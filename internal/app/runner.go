package app

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

const containerWorkRoot = "/app"
const claudeConfigTarget = "/home/agent/.claude.json"
const tmpfsMountOptions = "tmpfs-size=1m,tmpfs-mode=0555"

func buildArgs(appCfg Config, agent, language, project string, shell bool) []string {
	cfg := appCfg.Agents.byName(agent)
	projectPath := fmt.Sprintf("%s/%s/%s", containerWorkRoot, language, project)
	workdir := projectPath
	filteredTmpfsDirs := filterHostMountTargets(project, appCfg.Docker.TmpfsDirs)
	filteredNullFiles := filterHostMountTargets(project, appCfg.Docker.NullFiles)

	args := []string{
		"run", "--rm", "-it",
		"--add-host=" + appCfg.Docker.AddHost,
		"-v", cfg.StateMount,
	}

	args = appendEnvArg(args, "ALL_PROXY", appCfg.Docker.AllProxy)
	args = appendEnvArg(args, "HTTP_PROXY", appCfg.Docker.HTTPProxy)
	args = appendEnvArg(args, "HTTPS_PROXY", appCfg.Docker.HTTPSProxy)
	args = appendEnvArg(args, "NO_PROXY", appCfg.Docker.NoProxy)

	if cfg.hasConfigMount() {
		args = append(args, "-v", cfg.configMount(agent))
	}

	args = append(args, "-v", fmt.Sprintf("./%s:%s", project, projectPath))

	for _, target := range filteredTmpfsDirs.Items() {
		args = appendTmpfsMount(args, resolveMountTarget(projectPath, target))
	}

	for _, target := range filteredNullFiles.Items() {
		args = appendNullMount(args, resolveMountTarget(projectPath, target))
	}

	args = append(args,
		"-w", workdir,
		cfg.Image,
	)

	if shell {
		args = append(args, "bash")
	}

	return args
}

func run(args []string, dryRun bool) int {
	if dryRun {
		fmt.Println("docker " + strings.Join(args, " "))
		return 0
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	return 0
}

func appendEnvArg(args []string, key, value string) []string {
	if value == "" {
		return args
	}

	return append(args, "-e", key+"="+value)
}

func appendNullMount(args []string, target string) []string {
	return append(args, "-v", "/dev/null:"+target)
}

func appendTmpfsMount(args []string, target string) []string {
	return append(args, "--mount", fmt.Sprintf("type=tmpfs,destination=%s,%s", target, tmpfsMountOptions))
}

func resolveMountTarget(projectPath, target string) string {
	target = normalizeMountTarget(target)
	if path.IsAbs(target) {
		return target
	}

	return path.Join(projectPath, target)
}

func (c AgentConfig) hasConfigMount() bool {
	return c.ConfigSource != ""
}

func (c AgentConfig) configMount(agent string) string {
	switch agent {
	case "claude":
		return fmt.Sprintf("%s:%s:rw", c.ConfigSource, claudeConfigTarget)
	default:
		return c.ConfigSource
	}
}
