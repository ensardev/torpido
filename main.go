package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ensardev/torpido/internal/server"
	"github.com/ensardev/torpido/internal/ui"
)

func main() {
	// `torpido serve` starts the SSH server; plain `torpido` plays locally.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		addr := os.Getenv("TORPIDO_ADDR")
		if addr == "" {
			addr = ":2222"
		}
		if err := server.Run(addr); err != nil {
			fmt.Fprintln(os.Stderr, "torpido server error:", err)
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "torpido crashed:", err)
		os.Exit(1)
	}
}
