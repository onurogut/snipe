package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type processInfo struct {
	command  string
	filePath string
}

func parsePIDList(out string) ([]int, error) {
	seen := make(map[int]bool)
	var pids []int
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		pid, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			continue
		}
		if !seen[pid] {
			seen[pid] = true
			pids = append(pids, pid)
		}
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return pids, nil
}

var scriptExts = map[string]bool{
	".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".mjs": true, ".cjs": true,
	".py": true, ".go": true, ".rb": true, ".sh": true, ".pl": true, ".php": true,
	".rs": true, ".java": true, ".kt": true, ".swift": true,
}

func extractPath(cmdline string) string {
	parts := strings.Fields(cmdline)
	if len(parts) <= 1 {
		return parts[0]
	}

	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "-") {
			continue
		}
		if strings.Contains(p, "/") || strings.Contains(p, `\`) {
			return p
		}
	}

	for _, p := range parts[1:] {
		if strings.HasPrefix(p, "-") {
			continue
		}
		if scriptExts[filepath.Ext(p)] {
			return p
		}
	}

	return parts[0]
}

func looksLikeFile(s string) bool {
	if strings.ContainsAny(s, " '\"()") {
		return false
	}
	ext := filepath.Ext(s)
	return ext != "" && len(ext) <= 6
}
