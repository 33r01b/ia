package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const containerWorkRoot = "/app"

func buildArgs(appCfg Config, agent, language, project string) []string {
	cfg := appCfg.Agents.byName(agent)
	projectPath := fmt.Sprintf("%s/%s/%s", containerWorkRoot, language, project)

	projectMount := fmt.Sprintf("./%s:%s", project, projectPath)
	tmpfsMount := fmt.Sprintf("type=tmpfs,destination=%s/.idea,tmpfs-size=1m,tmpfs-mode=0555", projectPath)
	workdir := projectPath

	args := []string{
		"run", "--rm", "-it",
		"--add-host=" + appCfg.Docker.AddHost,
		"-e", "ALL_PROXY=" + appCfg.Docker.AllProxy,
		"-e", "HTTP_PROXY=" + appCfg.Docker.HTTPProxy,
		"-e", "HTTPS_PROXY=" + appCfg.Docker.HTTPSProxy,
		"-e", "NO_PROXY=" + appCfg.Docker.NoProxy,
		"-v", cfg.StateMount,
	}

	if cfg.hasConfigMount() {
		args = append(args, "-v", cfg.configMount())
	}

	args = append(args,
		"-v", projectMount,
		"--mount", tmpfsMount,
		"-w", workdir,
		cfg.Image,
	)

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

func (c AgentConfig) hasConfigMount() bool {
	return c.ConfigSource != "" && c.ConfigTarget != ""
}

func (c AgentConfig) configMount() string {
	return fmt.Sprintf("%s:%s:rw", c.ConfigSource, c.ConfigTarget)
}
