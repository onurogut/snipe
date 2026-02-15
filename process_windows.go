//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func checkDeps() error {
	return nil
}

func findPIDs(port string) ([]int, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	var pids []int
	seen := make(map[int]bool)
	suffix := ":" + port

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		local := fields[1]
		if !strings.HasSuffix(local, suffix) {
			continue
		}
		pid, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil || pid == 0 {
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

	// try wmic for full command line
	if out, err := exec.Command("wmic", "process", "where",
		"processid="+p, "get", "commandline", "/value").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "CommandLine=") {
				cmd := strings.TrimPrefix(line, "CommandLine=")
				if cmd != "" {
					info.command = cmd
					info.filePath = extractPath(cmd)
				}
				break
			}
		}
	}

	// fallback to tasklist for process name
	if info.command == "-" {
		if out, err := exec.Command("tasklist", "/FI",
			"PID eq "+p, "/FO", "CSV", "/NH").Output(); err == nil {
			line := strings.TrimSpace(string(out))
			if line != "" && !strings.Contains(line, "No tasks") {
				parts := strings.SplitN(line, ",", 3)
				if len(parts) >= 1 {
					name := strings.Trim(parts[0], "\"")
					info.command = name
				}
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

func getCwd(_ int) string {
	return ""
}

func forceKill(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if p.Kill() != nil {
		return false
	}
	return waitDead(pid, time.Second)
}

func gracefulKill(pid int) bool {
	// taskkill without /F sends WM_CLOSE for graceful shutdown
	if err := exec.Command("taskkill", "/PID", strconv.Itoa(pid)).Run(); err == nil {
		if waitDead(pid, 2*time.Second) {
			return true
		}
	}
	// force kill as fallback
	_ = exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
	return waitDead(pid, time.Second)
}

func waitDead(pid int, timeout time.Duration) bool {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		// can't open = already dead
		return true
	}
	defer windows.CloseHandle(h)

	ms := uint32(timeout.Milliseconds())
	r, _ := windows.WaitForSingleObject(h, ms)
	return r == windows.WAIT_OBJECT_0
}
