package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/torpido/internal/game"
	"github.com/ensardev/torpido/internal/lobby"
)

// botDelay is the pause before a bot fires back, so its move is readable.
const botDelay = 650 * time.Millisecond

// gamePhase is which screen of a match the player is looking at.
type gamePhase int

const (
	gameWaiting   gamePhase = iota // waiting for an opponent to join
	gamePlacing                    // placing your own fleet
	gamePlaceWait                  // you're done, waiting for the opponent to place
	gameBattle                     // firing back and forth
	gameOver                       // someone won
)

// roomUpdateMsg is delivered whenever the room signals a state change (the
// opponent joined, placed, or fired).
type roomUpdateMsg struct{}

// botMoveMsg triggers the bot's shot after a short, readable delay.
type botMoveMsg struct{}

// listenRoom blocks on the seat's update channel and reports back once. It is
// re-issued after every roomUpdateMsg so exactly one listener is ever pending.
func listenRoom(seat *lobby.Seat) tea.Cmd {
	return func() tea.Msg {
		<-seat.Updates()
		return roomUpdateMsg{}
	}
}

// gameModel is one player's view of a match in a room. Both bot games and
// human-vs-human games use it; the shared match lives in the room.
type gameModel struct {
	room     *lobby.Room
	seat     *lobby.Seat
	side     game.Side
	oppIsBot bool
	renderer *lipgloss.Renderer
	styles   styles

	phase gamePhase
	snap  lobby.Snapshot

	// placement state (local until sent to the room)
	fleet       []game.ShipType
	placeIndex  int
	cursor      game.Coord
	orientation game.Orientation

	// battle state
	aim     game.Coord
	message string
}

func newGameModel(room *lobby.Room, seat *lobby.Seat, r *lipgloss.Renderer) gameModel {
	side, _ := room.SideOf(seat)
	m := gameModel{
		room:        room,
		seat:        seat,
		side:        side,
		oppIsBot:    room.OpponentIsBot(side),
		renderer:    r,
		styles:      newStyles(r),
		fleet:       game.StandardFleet,
		orientation: game.Horizontal,
	}
	m.refresh()
	m.clampCursor()
	return m
}

func (m gameModel) Init() tea.Cmd { return listenRoom(m.seat) }

// refresh pulls a fresh snapshot and moves to the matching screen.
func (m *gameModel) refresh() {
	m.snap = m.room.Snapshot(m.side)
	switch {
	case m.snap.Over:
		m.phase = gameOver
	case !m.snap.OppPresent:
		m.phase = gameWaiting
	case m.snap.Phase == game.MatchBattle:
		m.phase = gameBattle
	case m.snap.YouPlaced:
		m.phase = gamePlaceWait
	default:
		m.phase = gamePlacing
	}
}

// maybeBotMove schedules the bot's shot when it is the bot's turn.
func (m gameModel) maybeBotMove() tea.Cmd {
	if m.oppIsBot && !m.snap.Over && m.snap.Phase == game.MatchBattle && !m.snap.YourTurn {
		return tea.Tick(botDelay, func(time.Time) tea.Msg { return botMoveMsg{} })
	}
	return nil
}

func (m gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case roomUpdateMsg:
		m.refresh()
		return m, tea.Batch(listenRoom(m.seat), m.maybeBotMove())
	case botMoveMsg:
		m.room.PlayBotTurn() // this notifies us, which drives the next refresh
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m gameModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	switch m.phase {
	case gamePlacing:
		return m.keyPlacing(msg)
	case gameBattle:
		return m.keyBattle(msg)
	case gameWaiting, gamePlaceWait:
		if msg.String() == "q" {
			return m, leaveGame
		}
	case gameOver:
		switch msg.String() {
		case "q", "enter", "r":
			return m, leaveGame
		}
	}
	return m, nil
}

func (m gameModel) keyPlacing(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, leaveGame
	case "up", "k":
		m.cursor.Row--
		m.clampCursor()
	case "down", "j":
		m.cursor.Row++
		m.clampCursor()
	case "left", "h":
		m.cursor.Col--
		m.clampCursor()
	case "right", "l":
		m.cursor.Col++
		m.clampCursor()
	case "r":
		if m.orientation == game.Horizontal {
			m.orientation = game.Vertical
		} else {
			m.orientation = game.Horizontal
		}
		m.clampCursor()
	case "enter", " ":
		st := m.fleet[m.placeIndex]
		coords := game.ShipCoords(m.cursor, st.Size, m.orientation)
		if m.room.PlaceShip(m.side, st.Name, st.Size, coords) {
			m.placeIndex++
			if m.placeIndex >= len(m.fleet) {
				m.room.FinishPlacing(m.side)
			} else {
				m.cursor = game.Coord{}
				m.orientation = game.Horizontal
			}
			m.refresh()
			m.clampCursor()
			return m, m.maybeBotMove()
		}
	}
	return m, nil
}

func (m gameModel) keyBattle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, leaveGame
	}
	if !m.snap.YourTurn {
		return m, nil // wait your turn
	}
	switch msg.String() {
	case "up", "k":
		if m.aim.Row > 0 {
			m.aim.Row--
		}
	case "down", "j":
		if m.aim.Row < game.BoardSize-1 {
			m.aim.Row++
		}
	case "left", "h":
		if m.aim.Col > 0 {
			m.aim.Col--
		}
	case "right", "l":
		if m.aim.Col < game.BoardSize-1 {
			m.aim.Col++
		}
	case "enter", " ":
		res, ship := m.room.Fire(m.side, m.aim)
		switch res {
		case game.FireInvalid:
			m.message = "Oraya zaten ateş ettin."
		case game.FireMiss:
			m.message = fmt.Sprintf("%s — ıska.", coordName(m.aim))
		case game.FireHit:
			m.message = fmt.Sprintf("%s — tam isabet! 💥", coordName(m.aim))
		case game.FireSunk:
			m.message = fmt.Sprintf("%s gemisini batırdın!", ship.Name)
		}
		m.refresh()
		return m, m.maybeBotMove()
	}
	return m, nil
}

// clampCursor keeps the placement cursor where the current ship still fits.
func (m *gameModel) clampCursor() {
	if m.placeIndex >= len(m.fleet) {
		return
	}
	st := m.fleet[m.placeIndex]
	maxRow, maxCol := game.BoardSize-1, game.BoardSize-1
	if m.orientation == game.Vertical {
		maxRow = game.BoardSize - st.Size
	} else {
		maxCol = game.BoardSize - st.Size
	}
	if m.cursor.Row < 0 {
		m.cursor.Row = 0
	}
	if m.cursor.Col < 0 {
		m.cursor.Col = 0
	}
	if m.cursor.Row > maxRow {
		m.cursor.Row = maxRow
	}
	if m.cursor.Col > maxCol {
		m.cursor.Col = maxCol
	}
}

// previewValid reports whether the current ship fits at the cursor, checked
// against the player's own board grid.
func (m gameModel) previewValid(coords []game.Coord) bool {
	for _, c := range coords {
		if !c.Valid() || m.snap.You[c.Row][c.Col] == game.CellShip {
			return false
		}
	}
	return true
}
