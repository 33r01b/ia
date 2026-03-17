package app

import (
	"slices"
	"strings"
)

func isValidLanguage(value string) bool {
	return isSafePathSegment(value)
}

func isValidProject(value string) bool {
	return isSafePathSegment(value)
}

func isSafePathSegment(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}

	return !strings.Contains(value, "/")
}

func sortedKeys(items map[string]bool) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}
