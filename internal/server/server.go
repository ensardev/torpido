// Package server serves torpido over SSH with Wish, so anyone can play by
// running `ssh host -p 2222` — no install, no signup.
//
// This is stage A of the networked build: every connection gets its own
// independent game against a bot. Shared lobby/rooms and human-vs-human come
// next; they will slot in by giving teaHandler access to shared state.
package server

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/ensardev/ssh-torpido/internal/lobby"
	"github.com/ensardev/ssh-torpido/internal/ui"
)

// hostKeyPath is where the server's SSH identity lives. It is generated on first
// run and is gitignored — never commit it.
const hostKeyPath = ".ssh/torpido_ed25519"

// Run starts the SSH server on addr (e.g. ":2222") and blocks until the process
// receives an interrupt, then shuts down gracefully.
func Run(addr string) error {
	// Make sure the directory for the generated host key exists.
	if err := os.MkdirAll(".ssh", 0o700); err != nil {
		return err
	}

	// One lobby is shared by every connection, so players meet in the same rooms.
	lb := lobby.New()

	srv, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			// Order matters: middleware runs bottom-to-top on the way in.
			bm.Middleware(teaHandler(lb)), // run the Bubble Tea app
			activeterm.Middleware(),       // reject connections without a real terminal
			logging.Middleware(),          // log who connects
		),
	)
	if err != nil {
		return err
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("torpido listening on %s — play with: ssh localhost -p %s", addr, port(addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalln("server error:", err)
		}
	}()

	<-done
	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

// teaHandler builds the lobby-backed app for each incoming SSH session. It gives
// the UI a renderer bound to *this* session's terminal, so colors match the
// player's terminal instead of the server's.
func teaHandler(lb *lobby.Lobby) bm.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		renderer := bm.MakeRenderer(s)
		return ui.NewRoot(lb, playerName(s), renderer), []tea.ProgramOption{tea.WithAltScreen()}
	}
}

// playerName is the display name for a connection, taken from the SSH username.
func playerName(s ssh.Session) string {
	if u := s.User(); u != "" && u != "root" {
		return u
	}
	return "denizci"
}

// port pulls the port out of a listen address for the friendly log line.
func port(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil && p != "" {
		return p
	}
	return addr
}
