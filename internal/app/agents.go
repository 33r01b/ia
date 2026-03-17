package app

import "strings"

var supportedAgents = map[string]bool{
	"claude": true,
	"codex":  true,
}

func isSupportedAgent(name string) bool {
	return supportedAgents[name]
}

func agentList() string {
	return strings.Join(sortedKeys(supportedAgents), ", ")
}

func (c AgentsConfig) byName(name string) AgentConfig {
	switch name {
	case "claude":
		return c.Claude
	case "codex":
		return c.Codex
	default:
		return AgentConfig{}
	}
}
