package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/niiithish/lazyports/internal/tui"
)

func main() {
	refresh := flag.Duration("refresh", 2*time.Second, "auto-refresh interval (0 to disable)")
	flag.Parse()

	if *refresh < 0 {
		fmt.Fprintln(os.Stderr, "refresh interval must be >= 0")
		os.Exit(1)
	}

	model := tui.New(*refresh)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
