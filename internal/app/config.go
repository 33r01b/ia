package app

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Docker DockerConfig
	Agents AgentsConfig
}

type DockerConfig struct {
	AllProxy   string
	HTTPProxy  string
	HTTPSProxy string
	NoProxy    string
	AddHost    string
	NullFiles  MountTargets
	TmpfsDirs  MountTargets
}

type MountTargets struct {
	items []string
	seen  map[string]struct{}
}

type AgentsConfig struct {
	Claude AgentConfig
	Codex  AgentConfig
}

type AgentConfig struct {
	Image        string
	StateMount   string
	ConfigSource string
}

type envConfig struct {
	ClaudeImage        string `env:"IA_CLAUDE_IMAGE" env-default:"claude-code"`
	ClaudeStateMount   string `env:"IA_CLAUDE_STATE_MOUNT" env-default:"claude_state:/home/agent/.claude"`
	ClaudeConfigSource string `env:"IA_CLAUDE_CONFIG_SOURCE"`
	CodexImage         string `env:"IA_CODEX_IMAGE" env-default:"codex-cli"`
	CodexStateMount    string `env:"IA_CODEX_STATE_MOUNT" env-default:"codex_state:/home/node/.codex"`
	CodexConfigSource  string `env:"IA_CODEX_CONFIG_SOURCE"`
	AllProxy           string `env:"IA_ALL_PROXY"`
	HTTPProxy          string `env:"IA_HTTP_PROXY"`
	HTTPSProxy         string `env:"IA_HTTPS_PROXY"`
	NoProxy            string `env:"IA_NO_PROXY" env-default:"host.docker.internal,localhost"`
	AddHost            string `env:"IA_DOCKER_ADD_HOST" env-default:"host.docker.internal:host-gateway"`
}

func loadConfig() (Config, error) {
	var envCfg envConfig

	if err := cleanenv.ReadEnv(&envCfg); err != nil {
		return Config{}, err
	}

	if envCfg.HTTPProxy == "" {
		envCfg.HTTPProxy = envCfg.AllProxy
	}

	if envCfg.HTTPSProxy == "" {
		envCfg.HTTPSProxy = envCfg.AllProxy
	}

	var nullFiles MountTargets
	var tmpfsDirs MountTargets
	tmpfsDirs.Add(".idea")

	cfg := Config{
		Docker: DockerConfig{
			AllProxy:   envCfg.AllProxy,
			HTTPProxy:  envCfg.HTTPProxy,
			HTTPSProxy: envCfg.HTTPSProxy,
			NoProxy:    envCfg.NoProxy,
			AddHost:    envCfg.AddHost,
			NullFiles:  nullFiles,
			TmpfsDirs:  tmpfsDirs,
		},
		Agents: AgentsConfig{
			Claude: AgentConfig{
				Image:        envCfg.ClaudeImage,
				StateMount:   envCfg.ClaudeStateMount,
				ConfigSource: envCfg.ClaudeConfigSource,
			},
			Codex: AgentConfig{
				Image:        envCfg.CodexImage,
				StateMount:   envCfg.CodexStateMount,
				ConfigSource: envCfg.CodexConfigSource,
			},
		},
	}

	return cfg, nil
}

func (c *Config) applyRunOptions(opts runOptions) {
	c.Docker.NullFiles.Merge(opts.nullFiles)
	c.Docker.TmpfsDirs.Merge(opts.tmpfsDirs)
	c.Docker.TmpfsDirs.Add(".idea")
}

func (c Config) validateForAgent(agentName string) error {
	if err := c.Docker.NullFiles.Validate(false); err != nil {
		return err
	}

	if err := c.Docker.TmpfsDirs.Validate(true); err != nil {
		return err
	}

	switch agentName {
	case "claude":
		if c.Agents.Claude.ConfigSource == "" {
			return fmt.Errorf("field %q is required but the value is not provided", "ClaudeConfigSource")
		}

		if _, err := os.Stat(c.Agents.Claude.ConfigSource); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("claude config file does not exist: %s", c.Agents.Claude.ConfigSource)
			}

			return fmt.Errorf("stat claude config file: %w", err)
		}
	}

	return nil
}

func parseList(raw string) []string {
	return parseMountTargets(raw).Items()
}

func parseMountTargets(raw string) MountTargets {
	var targets MountTargets
	if strings.TrimSpace(raw) == "" {
		return targets
	}

	for _, part := range strings.Split(raw, ",") {
		targets.Add(part)
	}

	return targets
}

func (m *MountTargets) Add(target string) {
	normalized := normalizeMountTarget(target)
	if normalized == "" {
		return
	}

	if m.seen == nil {
		m.seen = make(map[string]struct{})
	}

	if _, ok := m.seen[normalized]; ok {
		return
	}

	m.seen[normalized] = struct{}{}
	m.items = append(m.items, normalized)
}

func (m *MountTargets) Merge(other MountTargets) {
	for _, item := range other.items {
		m.Add(item)
	}
}

func (m MountTargets) Items() []string {
	return append([]string(nil), m.items...)
}

func (m MountTargets) Validate(allowDot bool) error {
	for _, target := range m.items {
		if target == "" {
			return fmt.Errorf("mount target must not be empty")
		}

		if !allowDot && target == "." {
			return fmt.Errorf("mount target %q must not point to the project root", target)
		}

		if path.IsAbs(target) {
			continue
		}

		if target == ".." || strings.HasPrefix(target, "../") {
			return fmt.Errorf("mount target %q must stay inside the project root or be absolute", target)
		}
	}

	return nil
}

func normalizeMountTarget(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}

	cleaned := path.Clean(target)
	if path.IsAbs(target) {
		return cleaned
	}

	return cleaned
}
