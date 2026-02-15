package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

var (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
	dim    = "\033[2m"
	bold   = "\033[1m"
)

func init() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		reset, red, green, yellow, cyan, dim, bold = "", "", "", "", "", "", ""
	}
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%serror:%s %s\n", red, reset, msg)
}

func printNotFound(port int) {
	fmt.Printf("%snothing on port %s%d%s\n", dim, cyan, port, reset)
}

func printKilled(port int, pid int, info processInfo) {
	fmt.Println(ruler())
	fmt.Printf("  %s%skill%s  :%s%d%s  pid %s%d%s\n",
		bold, red, reset, cyan, port, reset, yellow, pid, reset)
	fmt.Printf("  %scmd%s   %s\n", dim, reset, info.command)
	fmt.Printf("  %spath%s  %s\n", dim, reset, info.filePath)
	fmt.Println(ruler())
}

func printKillFailed(port int, pid int, info processInfo) {
	fmt.Println(ruler())
	fmt.Printf("  %s%sfail%s  :%s%d%s  pid %s%d%s — couldn't kill\n",
		bold, red, reset, cyan, port, reset, yellow, pid, reset)
	fmt.Printf("  %scmd%s   %s\n", dim, reset, info.command)
	fmt.Println(ruler())
}

func printDryRun(port int, pid int, info processInfo) {
	fmt.Println(ruler())
	fmt.Printf("  %s%sdry%s   :%s%d%s  pid %s%d%s — would kill\n",
		bold, yellow, reset, cyan, port, reset, yellow, pid, reset)
	fmt.Printf("  %scmd%s   %s\n", dim, reset, info.command)
	fmt.Printf("  %spath%s  %s\n", dim, reset, info.filePath)
	fmt.Println(ruler())
}

func printProcessInfo(port int, pid int, info processInfo) {
	fmt.Println(ruler())
	fmt.Printf("  %s%sfound%s :%s%d%s  pid %s%d%s\n",
		bold, green, reset, cyan, port, reset, yellow, pid, reset)
	fmt.Printf("  %scmd%s   %s\n", dim, reset, info.command)
	fmt.Printf("  %spath%s  %s\n", dim, reset, info.filePath)
	fmt.Println(ruler())
}

func ruler() string {
	return dim + "──────────────────────────────────────────────────" + reset
}
