// Package lobby holds torpido's server-side matchmaking: the rooms players sit
// in, the lobby that lists and creates them, and the reconcile loop that keeps
// a ready-to-play bot room waiting for every difficulty tier.
//
// Concurrency: the Lobby's mutex guards the set of rooms; each Room's mutex
// guards that room's seats and match. Locks are always taken lobby-first then
// room, never the other way, so the two never deadlock.
package lobby

import (
	"sync"

	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/players"
)

// Kind is whether a room hosts a human opponent or a waiting bot.
type Kind int

const (
	HumanRoom Kind = iota
	BotRoom
)

// Seat is one participant in a room. Exactly one of Human / bot is set.
type Seat struct {
	Name        string
	Human       bool
	Fingerprint string // persistent identity, for recording W/L (human seats)
	Wins        int    // career wins, for display (human seats)
	Losses      int    // career losses, for display (human seats)
	bot         *game.Bot
	updates     chan struct{} // signaled on state change (human seats only)
}

// NewHumanSeat makes a seat for a connected player.
func NewHumanSeat(name string) *Seat {
	return &Seat{Name: name, Human: true, updates: make(chan struct{}, 1)}
}

// Updates is the channel a human seat is signaled on when its room changes, so
// the UI can re-render when the opponent moves.
func (s *Seat) Updates() <-chan struct{} {
	return s.updates
}

// Room is one game slot: up to two seats and the match they play.
type Room struct {
	Code     string
	Kind     Kind
	Tier     game.Difficulty // meaningful for BotRoom
	Password string          // "" = open (HumanRoom only)
	Private  bool            // not listed publicly (HumanRoom only)

	mu          sync.Mutex
	match       *game.Match
	seats       [2]*Seat
	score       [2]int         // wins per side across rematches
	rematchWant [2]bool        // which sides asked for a rematch
	curScored   bool           // has the current finished match been counted yet?
	matchNo     int            // increments each rematch, so the UI can reset locals
	store       *players.Store // persistent W/L, shared from the lobby
}

// --- helpers below assume the room's mutex is held ---

func (r *Room) firstEmptyLocked() int {
	for i, s := range r.seats {
		if s == nil {
			return i
		}
	}
	return -1
}

func (r *Room) humanCountLocked() int {
	n := 0
	for _, s := range r.seats {
		if s != nil && s.Human {
			n++
		}
	}
	return n
}

// waitingLocked reports whether the room has an open seat and its match has not
// finished — i.e. it can still be joined.
func (r *Room) waitingLocked() bool {
	if r.firstEmptyLocked() < 0 {
		return false
	}
	_, over := r.match.Winner()
	return !over
}

// notifyLocked wakes every human seat so their UI re-renders. The send is
// non-blocking: the channel is buffered to depth 1 and a pending signal is
// enough (the UI reads the latest state, not a queue of events).
func (r *Room) notifyLocked() {
	for _, s := range r.seats {
		if s != nil && s.Human {
			select {
			case s.updates <- struct{}{}:
			default:
			}
		}
	}
}

// botIdle reports whether this is a bot room still waiting for a human.
func (r *Room) botIdle() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Kind == BotRoom && r.seats[0] == nil
}

// RoomInfo is a read-only snapshot of a room for the lobby list.
type RoomInfo struct {
	Code        string
	Kind        Kind
	Tier        game.Difficulty
	HasPassword bool
	HostName    string
	HostWins    int
	HostLosses  int
	Players     int
}
