package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/lobby"
	"github.com/ensardev/ssh-torpido/internal/players"
	"github.com/ensardev/ssh-torpido/internal/server"
	"github.com/ensardev/ssh-torpido/internal/ui"
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

	// Local play runs its own in-process lobby, so `go run .` shows the same
	// lobby-and-bots experience as connecting over SSH. There's no SSH key
	// locally, so the fingerprint is empty and stats aren't tracked.
	lb := lobby.New()
	store, _ := players.Open("stats.json")
	p := tea.NewProgram(ui.NewRoot(lb, "sen", "", store, lipgloss.DefaultRenderer()), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "torpido crashed:", err)
		os.Exit(1)
	}
}
