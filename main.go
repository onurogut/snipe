package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const version = "0.2.0"

type options struct {
	dryRun      bool
	listOnly    bool
	interactive bool
	quiet       bool
	graceful    bool
}

func main() {
	var opts options
	flag.BoolVar(&opts.dryRun, "d", false, "")
	flag.BoolVar(&opts.listOnly, "l", false, "")
	flag.BoolVar(&opts.interactive, "i", false, "")
	flag.BoolVar(&opts.quiet, "q", false, "")
	flag.BoolVar(&opts.graceful, "g", false, "")
	showVersion := flag.Bool("v", false, "")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Printf("snipe v%s\n", version)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	ports, err := parsePorts(args)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	if err := checkDeps(); err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	total := 0
	for _, port := range ports {
		total += run(port, opts)
	}

	if total == 0 && !opts.listOnly && !opts.dryRun {
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `snipe v%s — kill processes by port

usage: snipe [flags] <port> [port...]

  snipe 3000              kill whatever's on :3000
  snipe 3000 8080 4200    multiple ports
  snipe 3000-3005         port range

flags:
  -d    dry run, just show what would die
  -l    list processes without killing
  -i    ask before each kill
  -q    quiet, exit code only
  -g    try graceful shutdown first, force after 2s
  -v    version
`, version)
}

func parsePorts(args []string) ([]int, error) {
	var ports []int
	for _, arg := range args {
		if strings.Contains(arg, "-") && !strings.HasPrefix(arg, "-") {
			parts := strings.SplitN(arg, "-", 2)
			start, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", arg)
			}
			end, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", arg)
			}
			if start > end {
				return nil, fmt.Errorf("invalid range: %s", arg)
			}
			if end-start > 100 {
				return nil, fmt.Errorf("range too large (max 100): %s", arg)
			}
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			p, err := strconv.Atoi(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", arg)
			}
			if p < 1 || p > 65535 {
				return nil, fmt.Errorf("port out of range: %d", p)
			}
			ports = append(ports, p)
		}
	}
	return ports, nil
}

func run(port int, opts options) int {
	pids, err := findPIDs(strconv.Itoa(port))
	if err != nil || len(pids) == 0 {
		if !opts.quiet {
			printNotFound(port)
		}
		return 0
	}

	n := 0
	for _, pid := range pids {
		info := getProcessInfo(pid)

		if opts.listOnly {
			printProcessInfo(port, pid, info)
			n++
			continue
		}

		if opts.dryRun {
			printDryRun(port, pid, info)
			n++
			continue
		}

		if opts.interactive {
			printProcessInfo(port, pid, info)
			if !ask("  kill?") {
				continue
			}
		}

		ok := false
		if opts.graceful {
			ok = gracefulKill(pid)
		} else {
			ok = forceKill(pid)
		}

		if ok {
			n++
			if !opts.quiet {
				printKilled(port, pid, info)
			}
		} else if !opts.quiet {
			printKillFailed(port, pid, info)
		}
	}
	return n
}

func ask(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	s := bufio.NewScanner(os.Stdin)
	if s.Scan() {
		a := strings.TrimSpace(strings.ToLower(s.Text()))
		return a == "y" || a == "yes"
	}
	return false
}
