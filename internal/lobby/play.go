package lobby

import (
	"math/rand"
	"time"

	"github.com/ensardev/ssh-torpido/internal/game"
)

// This file is how the UI drives a match: every action goes through the room so
// the shared match stays consistent and the opponent is woken to re-render.

// Snapshot is a race-free, read-only view of a room's match from one side's
// point of view. Boards are value copies, so the UI can render them freely.
type Snapshot struct {
	Phase      game.MatchPhase
	You        [game.BoardSize][game.BoardSize]game.Cell // your board, ships shown
	Enemy      [game.BoardSize][game.BoardSize]game.Cell // enemy board, ships hidden
	EnemyFull  [game.BoardSize][game.BoardSize]game.Cell // enemy board revealed (for game over)
	YourTurn   bool
	YouPlaced  bool
	OppPlaced  bool
	OppPresent bool
	OppName    string
	Over       bool
	YouWon     bool
	Events     int          // total shots fired so far (for detecting new hits)
	Log        []game.Event // the most recent shots, oldest first
	YourScore  int          // wins across rematches, this player
	OppScore   int          // wins across rematches, the opponent
	RematchYou bool         // you asked for a rematch
	RematchOpp bool         // the opponent asked for a rematch
	MatchNo    int          // increments each rematch, so the UI can reset
}

// logTail is how many recent shots a snapshot carries for the battle log.
const logTail = 8

// SideOf returns which side a seat plays, and whether it is seated here at all.
func (r *Room) SideOf(seat *Seat) (game.Side, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, s := range r.seats {
		if s == seat {
			return game.Side(i), true
		}
	}
	return 0, false
}

// OpponentIsBot reports whether the seat opposite side is a bot.
func (r *Room) OpponentIsBot(side game.Side) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	opp := r.seats[side.Other()]
	return opp != nil && opp.bot != nil
}

// PlaceShip places one ship for side during the placing phase.
func (r *Room) PlaceShip(side game.Side, name string, size int, coords []game.Coord) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.match.Place(side, name, size, coords)
}

// FinishPlacing marks side ready and wakes the opponent so a waiting player sees
// the game start.
func (r *Room) FinishPlacing(side game.Side) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.match.FinishPlacing(side)
	r.notifyLocked()
}

// Fire fires for side at c, returns the result, and wakes the opponent on any
// real shot so their screen updates.
func (r *Room) Fire(side game.Side, c game.Coord) (game.FireResult, *game.Ship) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res, ship := r.match.Fire(side, c)
	if res != game.FireInvalid {
		r.scoreIfEndedLocked()
		r.notifyLocked()
	}
	return res, ship
}

// PlayBotTurn plays a single shot for the bot when it is the bot's turn, waking
// the human so they see it. It returns whether a shot was actually fired.
func (r *Room) PlayBotTurn() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.match.Phase() != game.MatchBattle {
		return false
	}
	turn := r.match.Turn()
	seat := r.seats[turn]
	if seat == nil || seat.bot == nil {
		return false
	}
	shot := seat.bot.NextShot()
	res, _ := r.match.Fire(turn, shot)
	seat.bot.Report(shot, res)
	r.scoreIfEndedLocked()
	r.notifyLocked()
	return res != game.FireInvalid
}

// Snapshot returns the current state of the match as seen by side.
func (r *Room) Snapshot(side game.Side) Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := r.match
	winner, over := m.Winner()
	opp := r.seats[side.Other()]
	snap := Snapshot{
		Phase:      m.Phase(),
		You:        m.Board(side).Grid(true),
		Enemy:      m.Board(side.Other()).Grid(false),
		EnemyFull:  m.Board(side.Other()).Grid(true),
		YourTurn:   m.Phase() == game.MatchBattle && m.Turn() == side,
		YouPlaced:  m.Placed(side),
		OppPlaced:  m.Placed(side.Other()),
		OppPresent: opp != nil,
		Over:       over,
		YouWon:     over && winner == side,
	}
	if opp != nil {
		snap.OppName = opp.Name
	}
	snap.YourScore = r.score[side]
	snap.OppScore = r.score[side.Other()]
	snap.RematchYou = r.rematchWant[side]
	snap.RematchOpp = r.rematchWant[side.Other()]
	snap.MatchNo = r.matchNo
	all := m.Events()
	snap.Events = len(all)
	if len(all) > logTail {
		all = all[len(all)-logTail:]
	}
	snap.Log = append([]game.Event(nil), all...) // copy so the UI never reads the live slice
	return snap
}

// scoreIfEndedLocked counts a win the first time the current match ends. Assumes
// the room's mutex is held.
func (r *Room) scoreIfEndedLocked() {
	if r.curScored {
		return
	}
	w, over := r.match.Winner()
	if !over {
		return
	}
	r.score[w]++
	r.curScored = true

	// Persist career W/L for human-vs-human matches only (bots don't count).
	if r.store == nil || r.Kind != HumanRoom {
		return
	}
	if ws := r.seats[w]; ws != nil && ws.Fingerprint != "" {
		r.store.RecordResult(ws.Fingerprint, true)
		ws.Wins++
	}
	if ls := r.seats[w.Other()]; ls != nil && ls.Fingerprint != "" {
		r.store.RecordResult(ls.Fingerprint, false)
		ls.Losses++
	}
}

// RequestRematch records that side wants to play again. A fresh match starts once
// both sides agree (a bot opponent always agrees).
func (r *Room) RequestRematch(side game.Side) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.match.Phase() != game.MatchOver {
		return
	}
	opp := r.seats[side.Other()]
	if opp == nil {
		return // no one left to rematch
	}
	r.rematchWant[side] = true
	botOpp := opp.bot != nil
	if botOpp || r.rematchWant[side.Other()] {
		r.resetMatchLocked()
	}
	r.notifyLocked()
}

// resetMatchLocked starts a fresh match, keeping the score. Assumes the room's
// mutex is held.
func (r *Room) resetMatchLocked() {
	r.match = game.NewMatch()
	r.rematchWant = [2]bool{}
	r.curScored = false
	r.matchNo++
	if r.Kind == BotRoom && r.seats[1] != nil {
		r.seats[1].bot = game.NewBot(r.Tier, time.Now().UnixNano())
		game.RandomPlacement(r.match.Board(game.SideB), game.StandardFleet, rand.New(rand.NewSource(time.Now().UnixNano())))
		r.match.FinishPlacing(game.SideB)
	}
}

// resetSeriesLocked returns the room to a fresh waiting state with the score
// wiped — used when a player is left alone so a new opponent can start over.
func (r *Room) resetSeriesLocked() {
	r.resetMatchLocked()
	r.score = [2]int{}
}
