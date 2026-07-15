package lobby

import "github.com/ensardev/torpido/internal/game"

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
}

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
	return snap
}
