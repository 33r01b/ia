package app

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Docker DockerConfig
	Agents AgentsConfig
}

type DockerConfig struct {
	AllProxy   string `env:"AGENT_ALL_PROXY" env-required:"true"`
	HTTPProxy  string `env:"AGENT_HTTP_PROXY"`
	HTTPSProxy string `env:"AGENT_HTTPS_PROXY"`
	NoProxy    string `env:"AGENT_NO_PROXY" env-default:"host.docker.internal,localhost"`
	AddHost    string `env:"AGENT_DOCKER_ADD_HOST" env-default:"host.docker.internal:host-gateway"`
}

type AgentsConfig struct {
	Claude AgentConfig
	Codex  AgentConfig
}

type AgentConfig struct {
	Image        string
	StateMount   string
	ConfigSource string
	ConfigTarget string
}

type envConfig struct {
	ClaudeImage        string `env:"AGENT_CLAUDE_IMAGE" env-default:"claude-code"`
	ClaudeStateMount   string `env:"AGENT_CLAUDE_STATE_MOUNT" env-default:"claude_state:/home/agent/.claude"`
	ClaudeConfigSource string `env:"AGENT_CLAUDE_CONFIG_SOURCE"`
	ClaudeConfigTarget string `env:"AGENT_CLAUDE_CONFIG_TARGET" env-default:"/home/agent/.claude.json"`
	CodexImage         string `env:"AGENT_CODEX_IMAGE" env-default:"codex-cli"`
	CodexStateMount    string `env:"AGENT_CODEX_STATE_MOUNT" env-default:"codex_state:/home/node/.codex"`
	CodexConfigSource  string `env:"AGENT_CODEX_CONFIG_SOURCE"`
	CodexConfigTarget  string `env:"AGENT_CODEX_CONFIG_TARGET"`
	DockerConfig
}

func loadConfig(agentName string) (Config, error) {
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

	cfg := Config{
		Docker: envCfg.DockerConfig,
		Agents: AgentsConfig{
			Claude: AgentConfig{
				Image:        envCfg.ClaudeImage,
				StateMount:   envCfg.ClaudeStateMount,
				ConfigSource: envCfg.ClaudeConfigSource,
				ConfigTarget: envCfg.ClaudeConfigTarget,
			},
			Codex: AgentConfig{
				Image:        envCfg.CodexImage,
				StateMount:   envCfg.CodexStateMount,
				ConfigSource: envCfg.CodexConfigSource,
				ConfigTarget: envCfg.CodexConfigTarget,
			},
		},
	}

	if err := cfg.validateForAgent(agentName); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validateForAgent(agentName string) error {
	switch agentName {
	case "claude":
		if c.Agents.Claude.ConfigSource == "" {
			return fmt.Errorf("field %q is required but the value is not provided", "ClaudeConfigSource")
		}
	}

	return nil
}
