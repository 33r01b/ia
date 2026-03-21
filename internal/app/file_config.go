package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type FileConfig struct {
	Project *string          `toml:"project"`
	Docker  FileDockerConfig `toml:"docker"`
}

type FileDockerConfig struct {
	AllProxy   *string  `toml:"all_proxy"`
	HTTPProxy  *string  `toml:"http_proxy"`
	HTTPSProxy *string  `toml:"https_proxy"`
	NoProxy    *string  `toml:"no_proxy"`
	AddHost    *string  `toml:"add_host"`
	MaskFiles  []string `toml:"mask_files"`
	MaskDirs   []string `toml:"mask_dirs"`
}

func configHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "ia"), nil
}

func globalConfigPath() (string, error) {
	base, err := configHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "config.toml"), nil
}

func projectConfigPath(project string) (string, error) {
	if !isValidProject(project) {
		return "", fmt.Errorf("invalid project %q", project)
	}

	base, err := configHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, "projects", project+".toml"), nil
}

func loadGlobalConfigFile() (FileConfig, bool, error) {
	configPath, err := globalConfigPath()
	if err != nil {
		return FileConfig{}, false, err
	}

	fileCfg, ok, err := loadOptionalFileConfig(configPath)
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("decode global config %s: %w", configPath, err)
	}

	return fileCfg, ok, nil
}

func loadProjectConfigFile(project string) (FileConfig, bool, error) {
	configPath, err := projectConfigPath(project)
	if err != nil {
		return FileConfig{}, false, err
	}

	fileCfg, ok, err := loadOptionalFileConfig(configPath)
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("decode project config %s: %w", configPath, err)
	}

	return fileCfg, ok, nil
}

func loadOptionalFileConfig(configPath string) (FileConfig, bool, error) {
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			return FileConfig{}, false, nil
		}
		return FileConfig{}, false, err
	}

	fileCfg, err := decodeFileConfig(configPath)
	if err != nil {
		return FileConfig{}, false, err
	}

	return fileCfg, true, nil
}

func decodeFileConfig(configPath string) (FileConfig, error) {
	var cfg FileConfig
	meta, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return FileConfig{}, err
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		parts := make([]string, 0, len(undecoded))
		for _, item := range undecoded {
			parts = append(parts, item.String())
		}
		return FileConfig{}, fmt.Errorf("unknown fields: %s", strings.Join(parts, ", "))
	}

	return cfg, nil
}

func validateProjectFileConfig(project string, cfg FileConfig) error {
	if cfg.Project != nil && *cfg.Project != project {
		return fmt.Errorf("project config mismatch: expected %q, got %q", project, *cfg.Project)
	}

	return nil
}
