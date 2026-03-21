package app

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadConfigDefaultsWithoutFiles(t *testing.T) {
	setTestHome(t)
	clearIAEnv(t)

	cfg, err := loadConfig("billing")
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if cfg.Docker.AddHost != "host.docker.internal:host-gateway" {
		t.Fatalf("unexpected AddHost: %q", cfg.Docker.AddHost)
	}
	if cfg.Docker.NoProxy != "host.docker.internal,localhost" {
		t.Fatalf("unexpected NoProxy: %q", cfg.Docker.NoProxy)
	}
	if cfg.Agents.Codex.Image != "codex-cli" {
		t.Fatalf("unexpected codex image: %q", cfg.Agents.Codex.Image)
	}
	if cfg.Agents.Claude.StateMount != "claude_state:/home/agent/.claude" {
		t.Fatalf("unexpected claude state mount: %q", cfg.Agents.Claude.StateMount)
	}
	if got := cfg.Docker.TmpfsDirs.Items(); !reflect.DeepEqual(got, []string{".idea"}) {
		t.Fatalf("unexpected tmpfs dirs: %#v", got)
	}
}

func TestLoadConfigMergesGlobalProjectAndEnv(t *testing.T) {
	homeDir := setTestHome(t)
	clearIAEnv(t)

	writeFile(t, filepath.Join(homeDir, ".config", "ia", "config.toml"), strings.TrimSpace(`
[docker]
all_proxy = "http://global-proxy:8080"
no_proxy = "global-no-proxy"
mask_files = [".env"]
mask_dirs = ["global-cache"]
`))

	writeFile(t, filepath.Join(homeDir, ".config", "ia", "projects", "billing.toml"), strings.TrimSpace(`
project = "billing"

[docker]
https_proxy = "https://project-proxy:8443"
no_proxy = "project-no-proxy"
mask_files = [".secrets/local.yaml"]
mask_dirs = ["project-cache"]
`))

	t.Setenv("IA_NO_PROXY", "env-no-proxy")
	t.Setenv("IA_HTTP_PROXY", "http://env-http:9090")

	cfg, err := loadConfig("billing")
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if cfg.Docker.AllProxy != "http://global-proxy:8080" {
		t.Fatalf("unexpected AllProxy: %q", cfg.Docker.AllProxy)
	}
	if cfg.Docker.HTTPProxy != "http://env-http:9090" {
		t.Fatalf("unexpected HTTPProxy: %q", cfg.Docker.HTTPProxy)
	}
	if cfg.Docker.HTTPSProxy != "https://project-proxy:8443" {
		t.Fatalf("unexpected HTTPSProxy: %q", cfg.Docker.HTTPSProxy)
	}
	if cfg.Docker.NoProxy != "env-no-proxy" {
		t.Fatalf("unexpected NoProxy: %q", cfg.Docker.NoProxy)
	}
	if got := cfg.Docker.NullFiles.Items(); !reflect.DeepEqual(got, []string{".env", ".secrets/local.yaml"}) {
		t.Fatalf("unexpected null files: %#v", got)
	}
	if got := cfg.Docker.TmpfsDirs.Items(); !reflect.DeepEqual(got, []string{"global-cache", "project-cache", ".idea"}) {
		t.Fatalf("unexpected tmpfs dirs: %#v", got)
	}
}

func TestLoadConfigFallsBackToAllProxyAfterMerge(t *testing.T) {
	homeDir := setTestHome(t)
	clearIAEnv(t)

	writeFile(t, filepath.Join(homeDir, ".config", "ia", "config.toml"), strings.TrimSpace(`
[docker]
all_proxy = "http://global-proxy:8080"
`))

	cfg, err := loadConfig("billing")
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if cfg.Docker.HTTPProxy != "http://global-proxy:8080" {
		t.Fatalf("unexpected HTTPProxy: %q", cfg.Docker.HTTPProxy)
	}
	if cfg.Docker.HTTPSProxy != "http://global-proxy:8080" {
		t.Fatalf("unexpected HTTPSProxy: %q", cfg.Docker.HTTPSProxy)
	}
}

func TestLoadConfigRejectsProjectMismatch(t *testing.T) {
	homeDir := setTestHome(t)
	clearIAEnv(t)

	writeFile(t, filepath.Join(homeDir, ".config", "ia", "projects", "billing.toml"), strings.TrimSpace(`
project = "other"
`))

	_, err := loadConfig("billing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), `project config mismatch: expected "billing", got "other"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadConfigRejectsUnknownTomlField(t *testing.T) {
	homeDir := setTestHome(t)
	clearIAEnv(t)

	writeFile(t, filepath.Join(homeDir, ".config", "ia", "config.toml"), strings.TrimSpace(`
[docker]
unknown_field = "value"
`))

	_, err := loadConfig("billing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown fields: docker.unknown_field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func setTestHome(t *testing.T) string {
	t.Helper()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	return homeDir
}

func clearIAEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"IA_CLAUDE_IMAGE",
		"IA_CLAUDE_STATE_MOUNT",
		"IA_CLAUDE_CONFIG_SOURCE",
		"IA_CODEX_IMAGE",
		"IA_CODEX_STATE_MOUNT",
		"IA_CODEX_CONFIG_SOURCE",
		"IA_ALL_PROXY",
		"IA_HTTP_PROXY",
		"IA_HTTPS_PROXY",
		"IA_NO_PROXY",
		"IA_DOCKER_ADD_HOST",
	} {
		t.Setenv(key, "")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
