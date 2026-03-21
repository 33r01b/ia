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
	ClaudeImage        string `env:"IA_CLAUDE_IMAGE"`
	ClaudeStateMount   string `env:"IA_CLAUDE_STATE_MOUNT"`
	ClaudeConfigSource string `env:"IA_CLAUDE_CONFIG_SOURCE"`
	CodexImage         string `env:"IA_CODEX_IMAGE"`
	CodexStateMount    string `env:"IA_CODEX_STATE_MOUNT"`
	CodexConfigSource  string `env:"IA_CODEX_CONFIG_SOURCE"`
	AllProxy           string `env:"IA_ALL_PROXY"`
	HTTPProxy          string `env:"IA_HTTP_PROXY"`
	HTTPSProxy         string `env:"IA_HTTPS_PROXY"`
	NoProxy            string `env:"IA_NO_PROXY"`
	AddHost            string `env:"IA_DOCKER_ADD_HOST"`
}

func loadConfig(project string) (Config, error) {
	cfg := defaultConfig()

	globalFileCfg, ok, err := loadGlobalConfigFile()
	if err != nil {
		return Config{}, err
	}
	if ok {
		cfg.applyFileConfig(globalFileCfg)
	}

	projectFileCfg, ok, err := loadProjectConfigFile(project)
	if err != nil {
		return Config{}, err
	}
	if ok {
		if err := validateProjectFileConfig(project, projectFileCfg); err != nil {
			return Config{}, err
		}
		cfg.applyFileConfig(projectFileCfg)
	}

	envCfg, err := envConfigOverrides()
	if err != nil {
		return Config{}, err
	}
	applyEnvConfig(&cfg, envCfg)
	finalizeConfig(&cfg)

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Docker: DockerConfig{
			NoProxy: "host.docker.internal,localhost",
			AddHost: "host.docker.internal:host-gateway",
		},
		Agents: AgentsConfig{
			Claude: AgentConfig{
				Image:      "claude-code",
				StateMount: "claude_state:/home/agent/.claude",
			},
			Codex: AgentConfig{
				Image:      "codex-cli",
				StateMount: "codex_state:/home/node/.codex",
			},
		},
	}
}

func envConfigOverrides() (envConfig, error) {
	var envCfg envConfig
	if err := cleanenv.ReadEnv(&envCfg); err != nil {
		return envConfig{}, err
	}
	return envCfg, nil
}

func applyEnvConfig(cfg *Config, envCfg envConfig) {
	if envCfg.AllProxy != "" {
		cfg.Docker.AllProxy = envCfg.AllProxy
	}
	if envCfg.HTTPProxy != "" {
		cfg.Docker.HTTPProxy = envCfg.HTTPProxy
	}
	if envCfg.HTTPSProxy != "" {
		cfg.Docker.HTTPSProxy = envCfg.HTTPSProxy
	}
	if envCfg.NoProxy != "" {
		cfg.Docker.NoProxy = envCfg.NoProxy
	}
	if envCfg.AddHost != "" {
		cfg.Docker.AddHost = envCfg.AddHost
	}

	if envCfg.ClaudeImage != "" {
		cfg.Agents.Claude.Image = envCfg.ClaudeImage
	}
	if envCfg.ClaudeStateMount != "" {
		cfg.Agents.Claude.StateMount = envCfg.ClaudeStateMount
	}
	if envCfg.ClaudeConfigSource != "" {
		cfg.Agents.Claude.ConfigSource = envCfg.ClaudeConfigSource
	}

	if envCfg.CodexImage != "" {
		cfg.Agents.Codex.Image = envCfg.CodexImage
	}
	if envCfg.CodexStateMount != "" {
		cfg.Agents.Codex.StateMount = envCfg.CodexStateMount
	}
	if envCfg.CodexConfigSource != "" {
		cfg.Agents.Codex.ConfigSource = envCfg.CodexConfigSource
	}
}

func finalizeConfig(cfg *Config) {
	if cfg.Docker.HTTPProxy == "" {
		cfg.Docker.HTTPProxy = cfg.Docker.AllProxy
	}

	if cfg.Docker.HTTPSProxy == "" {
		cfg.Docker.HTTPSProxy = cfg.Docker.AllProxy
	}

	cfg.Docker.TmpfsDirs.Add(".idea")
}

func (c *Config) applyFileConfig(fileCfg FileConfig) {
	c.applyFileDockerConfig(fileCfg.Docker)
}

func (c *Config) applyFileDockerConfig(fileCfg FileDockerConfig) {
	if fileCfg.AllProxy != nil {
		c.Docker.AllProxy = *fileCfg.AllProxy
	}
	if fileCfg.HTTPProxy != nil {
		c.Docker.HTTPProxy = *fileCfg.HTTPProxy
	}
	if fileCfg.HTTPSProxy != nil {
		c.Docker.HTTPSProxy = *fileCfg.HTTPSProxy
	}
	if fileCfg.NoProxy != nil {
		c.Docker.NoProxy = *fileCfg.NoProxy
	}
	if fileCfg.AddHost != nil {
		c.Docker.AddHost = *fileCfg.AddHost
	}

	for _, target := range fileCfg.MaskFiles {
		c.Docker.NullFiles.Add(target)
	}
	for _, target := range fileCfg.MaskDirs {
		c.Docker.TmpfsDirs.Add(target)
	}
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
