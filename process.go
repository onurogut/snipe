package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type processInfo struct {
	command  string
	filePath string
}

func checkDeps() error {
	if _, err := exec.LookPath("lsof"); err != nil {
		if runtime.GOOS == "linux" {
			if _, err := exec.LookPath("ss"); err != nil {
				return fmt.Errorf("lsof or ss required, install one: apt install lsof")
			}
			return nil
		}
		return fmt.Errorf("lsof not found")
	}
	return nil
}

func findPIDs(port string) ([]int, error) {
	pids, err := findPIDsLsof(port)
	if err == nil && len(pids) > 0 {
		return pids, nil
	}
	if runtime.GOOS == "linux" {
		return findPIDsLinux(port)
	}
	return nil, err
}

func findPIDsLsof(port string) ([]int, error) {
	out, err := exec.Command("lsof", "-t", "-i:"+port).Output()
	if err != nil {
		return nil, err
	}
	return parsePIDList(string(out))
}

func findPIDsLinux(port string) ([]int, error) {
	pids, err := findPIDsSs(port)
	if err == nil && len(pids) > 0 {
		return pids, nil
	}
	return findPIDsProc(port)
}

func findPIDsSs(port string) ([]int, error) {
	out, err := exec.Command("ss", "-tlnp", "sport", "=", ":"+port).Output()
	if err != nil {
		return nil, err
	}

	var pids []int
	seen := make(map[int]bool)
	for _, line := range strings.Split(string(out), "\n") {
		idx := strings.Index(line, "pid=")
		if idx == -1 {
			continue
		}
		rest := line[idx+4:]
		end := strings.IndexAny(rest, ",) ")
		if end == -1 {
			end = len(rest)
		}
		pid, err := strconv.Atoi(rest[:end])
		if err != nil {
			continue
		}
		if !seen[pid] {
			seen[pid] = true
			pids = append(pids, pid)
		}
	}
	return pids, nil
}

func findPIDsProc(port string) ([]int, error) {
	portNum, _ := strconv.Atoi(port)
	hexPort := fmt.Sprintf("%04X", portNum)

	data, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		data, err = os.ReadFile("/proc/net/tcp6")
		if err != nil {
			return nil, fmt.Errorf("/proc/net/tcp not available")
		}
	}

	var inodes []string
	for _, line := range strings.Split(string(data), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		parts := strings.Split(fields[1], ":")
		if len(parts) == 2 && strings.EqualFold(parts[1], hexPort) {
			inodes = append(inodes, fields[9])
		}
	}

	if len(inodes) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return mapInodesToPIDs(inodes)
}

func mapInodesToPIDs(inodes []string) ([]int, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	set := make(map[string]bool, len(inodes))
	for _, n := range inodes {
		set[n] = true
	}

	var pids []int
	seen := make(map[int]bool)
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join("/proc", entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if strings.HasPrefix(link, "socket:[") {
				inode := link[8 : len(link)-1]
				if set[inode] && !seen[pid] {
					seen[pid] = true
					pids = append(pids, pid)
				}
			}
		}
	}
	return pids, nil
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

func getProcessInfo(pid int) processInfo {
	info := processInfo{command: "-", filePath: "-"}
	p := strconv.Itoa(pid)

	// linux: read from procfs directly
	if runtime.GOOS == "linux" {
		if raw, err := os.ReadFile("/proc/" + p + "/cmdline"); err == nil {
			cmd := strings.TrimSpace(strings.ReplaceAll(string(raw), "\x00", " "))
			if cmd != "" {
				info.command = cmd
				info.filePath = extractPath(cmd)
			}
		}
	}

	if info.command == "-" {
		if out, err := exec.Command("ps", "-p", p, "-o", "args=").Output(); err == nil {
			cmd := strings.TrimSpace(string(out))
			if cmd != "" {
				info.command = cmd
				info.filePath = extractPath(cmd)
			}
		}
	}

	if info.filePath != "-" && !filepath.IsAbs(info.filePath) && looksLikeFile(info.filePath) {
		if cwd := getCwd(pid); cwd != "" {
			info.filePath = filepath.Join(cwd, info.filePath)
		}
	}

	return info
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
		if strings.Contains(p, "/") {
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

func getCwd(pid int) string {
	p := strconv.Itoa(pid)

	if runtime.GOOS == "linux" {
		if target, err := os.Readlink("/proc/" + p + "/cwd"); err == nil {
			return target
		}
	}

	out, err := exec.Command("lsof", "-p", p, "-Fn").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if line == "fcwd" && i+1 < len(lines) && strings.HasPrefix(lines[i+1], "n") {
			return lines[i+1][1:]
		}
	}
	return ""
}

func forceKill(pid int) bool {
	if syscall.Kill(pid, syscall.SIGKILL) != nil {
		return false
	}
	return waitDead(pid, time.Second)
}

func gracefulKill(pid int) bool {
	if syscall.Kill(pid, syscall.SIGTERM) != nil {
		return false
	}
	if waitDead(pid, 2*time.Second) {
		return true
	}
	// still hanging around, nuke it
	syscall.Kill(pid, syscall.SIGKILL)
	return waitDead(pid, time.Second)
}

func waitDead(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, 0) != nil {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
