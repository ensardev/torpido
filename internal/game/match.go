package game

// Match is a two-player game: two boards, simultaneous fleet placement, then
// alternating fire until one fleet is sunk. It is pure logic with no locking —
// the room that owns a Match is responsible for serializing access to it.
//
// The two participants are called sides. A side may be driven by a human over
// SSH or by a bot; Match does not care which, it only tracks the rules.
type Match struct {
	boards    [2]*Board
	placed    [2]bool // has this side finished placing its fleet?
	turn      Side    // whose turn it is to fire (battle phase)
	phase     MatchPhase
	winner    Side
	hasWinner bool
}

// Side identifies one of the two players.
type Side int

const (
	SideA Side = iota
	SideB
)

// Other returns the opposing side.
func (s Side) Other() Side {
	if s == SideA {
		return SideB
	}
	return SideA
}

// MatchPhase is the stage a match is in.
type MatchPhase int

const (
	MatchPlacing MatchPhase = iota // both sides placing fleets
	MatchBattle                    // alternating fire
	MatchOver                      // someone has won
)

// NewMatch returns a match with two empty boards, ready for placement.
func NewMatch() *Match {
	return &Match{
		boards: [2]*Board{NewBoard(), NewBoard()},
		phase:  MatchPlacing,
	}
}

// Board returns the board belonging to side s (the one holding s's ships).
func (m *Match) Board(s Side) *Board { return m.boards[s] }

// Phase reports the current stage of the match.
func (m *Match) Phase() MatchPhase { return m.phase }

// Turn reports whose turn it is to fire. Only meaningful during MatchBattle.
func (m *Match) Turn() Side { return m.turn }

// Placed reports whether side s has finished placing its fleet.
func (m *Match) Placed(s Side) bool { return m.placed[s] }

// Winner returns the winning side and whether the match has ended.
func (m *Match) Winner() (Side, bool) { return m.winner, m.hasWinner }

// Place puts one ship on side s's board during the placing phase. It returns
// false if placing is over, the side already finished, or the squares are taken.
func (m *Match) Place(s Side, name string, size int, coords []Coord) bool {
	if m.phase != MatchPlacing || m.placed[s] {
		return false
	}
	return m.boards[s].Place(name, size, coords)
}

// FinishPlacing marks side s ready. Once both sides are ready the battle starts
// with SideA firing first.
func (m *Match) FinishPlacing(s Side) {
	if m.phase != MatchPlacing {
		return
	}
	m.placed[s] = true
	if m.placed[SideA] && m.placed[SideB] {
		m.phase = MatchBattle
		m.turn = SideA
	}
}

// Fire resolves a shot by side s at coord c on the opponent's board. It is only
// legal on s's turn during battle; otherwise it returns FireInvalid and nothing
// changes. A hit keeps the rules simple by still passing the turn (strict
// one-shot alternation). Sinking the opponent's last ship ends the match.
func (m *Match) Fire(s Side, c Coord) (FireResult, *Ship) {
	if m.phase != MatchBattle || m.turn != s {
		return FireInvalid, nil
	}
	target := m.boards[s.Other()]
	res, ship := target.Fire(c)
	if res == FireInvalid {
		return res, nil
	}
	if target.AllSunk() {
		m.phase = MatchOver
		m.winner = s
		m.hasWinner = true
		return res, ship
	}
	m.turn = s.Other()
	return res, ship
}

// Resign ends the match immediately with the other side winning. Used when a
// player disconnects or quits mid-game.
func (m *Match) Resign(s Side) {
	if m.phase == MatchOver {
		return
	}
	m.phase = MatchOver
	m.winner = s.Other()
	m.hasWinner = true
}
