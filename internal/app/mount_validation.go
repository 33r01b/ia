package app

import (
	"os"
	"path"
	"path/filepath"
)

func filterHostMountTargets(project string, targets MountTargets) MountTargets {
	projectRoot := filepath.Join(".", project)
	var filtered MountTargets

	for _, target := range targets.Items() {
		if shouldMountTarget(projectRoot, target) {
			filtered.Add(target)
		}
	}

	return filtered
}

func shouldMountTarget(projectRoot, target string) bool {
	if path.IsAbs(target) {
		return true
	}

	hostPath := filepath.Join(projectRoot, filepath.FromSlash(target))
	if _, err := os.Stat(hostPath); err != nil {
		return false
	}

	return true
}
