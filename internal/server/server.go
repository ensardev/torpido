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

	"github.com/ensardev/torpido/internal/ui"
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

	srv, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(
			// Order matters: middleware runs bottom-to-top on the way in.
			bm.Middleware(teaHandler), // run the Bubble Tea game
			activeterm.Middleware(),   // reject connections without a real terminal
			logging.Middleware(),      // log who connects
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

// teaHandler builds a fresh game for each incoming SSH session. Wish's Bubble
// Tea middleware wires the session's terminal (size, colors, input) into the
// program for us, so the same Model that runs locally runs over SSH unchanged.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return ui.NewModel(), []tea.ProgramOption{tea.WithAltScreen()}
}

// port pulls the port out of a listen address for the friendly log line.
func port(addr string) string {
	if _, p, err := net.SplitHostPort(addr); err == nil && p != "" {
		return p
	}
	return addr
}
