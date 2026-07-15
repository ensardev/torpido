package lobby

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/ensardev/torpido/internal/game"
)

// Errors returned when joining a room.
var (
	ErrNoRoom      = errors.New("böyle bir oda yok")
	ErrRoomFull    = errors.New("oda dolu")
	ErrBadPassword = errors.New("şifre yanlış")
)

// botTiers are the difficulties that always have a ready bot room waiting.
var botTiers = []game.Difficulty{game.Rookie, game.Admiral, game.SeaWolf}

const (
	codeLen      = 4
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ" // no I/O/1/0 to avoid confusion
)

// Lobby is the set of all rooms plus the rule that keeps one idle bot room
// waiting per difficulty tier.
type Lobby struct {
	mu    sync.Mutex
	rooms map[string]*Room
	rng   *rand.Rand
}

// New builds a lobby with the standard three bot rooms already waiting.
func New() *Lobby {
	l := &Lobby{
		rooms: make(map[string]*Room),
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	l.mu.Lock()
	l.reconcileLocked()
	l.mu.Unlock()
	return l
}

// CreateRoom opens a new human room with the creator seated, returning it. An
// empty password means the room is open; private rooms are joinable by code but
// hidden from the public list.
func (l *Lobby) CreateRoom(creator *Seat, password string, private bool) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()
	code := l.uniqueCodeLocked()
	r := &Room{
		Code:     code,
		Kind:     HumanRoom,
		Password: password,
		Private:  private,
		match:    game.NewMatch(),
	}
	r.seats[0] = creator
	l.rooms[code] = r
	return r
}

// JoinByCode seats joiner in the room with the given code. Bots rooms ignore the
// password. Joining a bot room triggers a reconcile so a fresh idle one appears.
func (l *Lobby) JoinByCode(code string, joiner *Seat, password string) (*Room, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	r, ok := l.rooms[strings.ToUpper(strings.TrimSpace(code))]
	if !ok {
		return nil, ErrNoRoom
	}

	r.mu.Lock()
	if r.Kind == HumanRoom && r.Password != "" && r.Password != password {
		r.mu.Unlock()
		return nil, ErrBadPassword
	}
	idx := r.firstEmptyLocked()
	if idx < 0 || !r.waitingLocked() {
		r.mu.Unlock()
		return nil, ErrRoomFull
	}
	r.seats[idx] = joiner
	r.notifyLocked()
	isBot := r.Kind == BotRoom
	r.mu.Unlock()

	if isBot {
		l.reconcileLocked()
	}
	return r, nil
}

// QuickMatch drops joiner into any open public human room, or opens a new one
// for them to wait in if none is available.
func (l *Lobby) QuickMatch(joiner *Seat) *Room {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, r := range l.rooms {
		r.mu.Lock()
		open := r.Kind == HumanRoom && !r.Private && r.Password == "" && r.waitingLocked()
		if open {
			idx := r.firstEmptyLocked()
			r.seats[idx] = joiner
			r.notifyLocked()
			r.mu.Unlock()
			return r
		}
		r.mu.Unlock()
	}

	code := l.uniqueCodeLocked()
	r := &Room{Code: code, Kind: HumanRoom, match: game.NewMatch()}
	r.seats[0] = joiner
	l.rooms[code] = r
	return r
}

// Leave removes seat from the room. A departing player forfeits an unfinished
// match to the opponent. Emptied human rooms and any bot room the player leaves
// are removed, then the bot invariant is restored.
func (l *Lobby) Leave(r *Room, seat *Seat) {
	l.mu.Lock()
	defer l.mu.Unlock()

	r.mu.Lock()
	leaver, found := -1, false
	for i, s := range r.seats {
		if s == seat {
			r.seats[i], leaver, found = nil, i, true
		}
	}
	if found {
		if _, over := r.match.Winner(); !over {
			r.match.Resign(game.Side(leaver)) // opponent wins the forfeit
		}
	}
	humans := r.humanCountLocked()
	isBot := r.Kind == BotRoom
	r.notifyLocked()
	r.mu.Unlock()

	if isBot || humans == 0 {
		delete(l.rooms, r.Code)
	}
	l.reconcileLocked()
}

// PublicRooms returns the joinable rooms for the lobby list: every idle bot room
// plus open, non-private human rooms still waiting for a second player.
func (l *Lobby) PublicRooms() []RoomInfo {
	l.mu.Lock()
	defer l.mu.Unlock()

	var out []RoomInfo
	for _, r := range l.rooms {
		r.mu.Lock()
		switch {
		case r.Kind == BotRoom && r.seats[0] == nil:
			out = append(out, RoomInfo{
				Code: r.Code, Kind: BotRoom, Tier: r.Tier,
				HostName: r.Tier.Name(), Players: 1,
			})
		case r.Kind == HumanRoom && !r.Private && r.waitingLocked():
			host := ""
			if r.seats[0] != nil {
				host = r.seats[0].Name
			}
			out = append(out, RoomInfo{
				Code: r.Code, Kind: HumanRoom,
				HasPassword: r.Password != "", HostName: host, Players: 1,
			})
		}
		r.mu.Unlock()
	}
	return out
}

// reconcileLocked restores the invariant: exactly one idle bot room per tier.
// It runs after every join/leave, so the lobby self-heals no matter the order of
// events (the same idea as a Kubernetes reconcile loop). Assumes l.mu is held.
func (l *Lobby) reconcileLocked() {
	for _, tier := range botTiers {
		var idle []*Room
		for _, r := range l.rooms {
			if r.Kind == BotRoom && r.Tier == tier && r.botIdle() {
				idle = append(idle, r)
			}
		}
		switch {
		case len(idle) == 0:
			l.createBotRoomLocked(tier)
		case len(idle) > 1:
			for _, extra := range idle[1:] {
				delete(l.rooms, extra.Code)
			}
		}
	}
}

// createBotRoomLocked adds a fresh idle bot room whose bot has already placed its
// fleet, waiting for a human to take the open seat. Assumes l.mu is held.
func (l *Lobby) createBotRoomLocked(tier game.Difficulty) *Room {
	code := l.uniqueCodeLocked()
	r := &Room{Code: code, Kind: BotRoom, Tier: tier, match: game.NewMatch()}
	r.seats[1] = &Seat{Name: tier.Name(), bot: game.NewBot(tier, l.rng.Int63())}
	game.RandomPlacement(r.match.Board(game.SideB), game.StandardFleet, rand.New(rand.NewSource(l.rng.Int63())))
	r.match.FinishPlacing(game.SideB)
	l.rooms[code] = r
	return r
}

// uniqueCodeLocked returns a room code not already in use. Assumes l.mu is held.
func (l *Lobby) uniqueCodeLocked() string {
	for {
		b := make([]byte, codeLen)
		for i := range b {
			b[i] = codeAlphabet[l.rng.Intn(len(codeAlphabet))]
		}
		code := string(b)
		if _, exists := l.rooms[code]; !exists {
			return code
		}
	}
}
