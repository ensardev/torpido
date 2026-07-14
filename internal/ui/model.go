package ui

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ensardev/torpido/internal/game"
)

// phase is which screen of the game we are on.
type phase int

const (
	phasePlacement phase = iota
	phaseBattle
	phaseGameOver
)

// botDelay is the pause before the bot fires back, so its move is readable.
const botDelay = 650 * time.Millisecond

// botFireMsg is delivered after botDelay to trigger the bot's shot.
type botFireMsg struct{}

// Model is the whole game state for one local session.
type Model struct {
	phase phase

	player *game.Board // your grid (ships visible to you)
	enemy  *game.Board // the bot's grid (ships hidden)
	bot    *game.Bot

	// placement phase
	fleet       []game.ShipType
	placeIndex  int
	cursor      game.Coord
	orientation game.Orientation

	// battle phase
	aim     game.Coord
	waiting bool // true while the bot is "thinking"; input is ignored
	message string

	// game over
	playerWon bool

	rng *rand.Rand
}

// NewModel returns a fresh game sitting on the placement screen.
func NewModel() Model {
	m := Model{
		phase:       phasePlacement,
		player:      game.NewBoard(),
		fleet:       game.StandardFleet,
		orientation: game.Horizontal,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	m.clampCursor()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

// clampCursor keeps the placement cursor where the current ship still fits on
// the board, so the preview is never partly off the grid.
func (m *Model) clampCursor() {
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

func botThink() tea.Cmd {
	return tea.Tick(botDelay, func(time.Time) tea.Msg { return botFireMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case botFireMsg:
		return m.botFire()
	case tea.KeyMsg:
		switch m.phase {
		case phasePlacement:
			return m.updatePlacement(msg)
		case phaseBattle:
			return m.updateBattle(msg)
		case phaseGameOver:
			return m.updateGameOver(msg)
		}
	}
	return m, nil
}

func (m Model) updatePlacement(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
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
		if m.player.Place(st.Name, st.Size, coords) {
			m.placeIndex++
			if m.placeIndex >= len(m.fleet) {
				m.startBattle()
			} else {
				m.cursor = game.Coord{}
				m.orientation = game.Horizontal
				m.clampCursor()
			}
		}
	}
	return m, nil
}

func (m *Model) startBattle() {
	m.enemy = game.NewBoard()
	game.RandomPlacement(m.enemy, m.fleet, m.rng)
	m.bot = game.NewBot(time.Now().UnixNano())
	m.phase = phaseBattle
	m.aim = game.Coord{}
	m.message = "Muharebe başlasın! İlk atışını yap."
}

func (m Model) updateBattle(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.waiting {
		if s := msg.String(); s == "ctrl+c" || s == "q" {
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
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
		result, ship := m.enemy.Fire(m.aim)
		switch result {
		case game.FireInvalid:
			m.message = "Oraya zaten ateş ettin."
		case game.FireMiss:
			m.message = fmt.Sprintf("%s — sulara düştü, ıska.", coordName(m.aim))
			m.waiting = true
			return m, botThink()
		case game.FireHit:
			m.message = fmt.Sprintf("%s — tam isabet! 💥", coordName(m.aim))
			m.waiting = true
			return m, botThink()
		case game.FireSunk:
			if m.enemy.AllSunk() {
				m.phase = phaseGameOver
				m.playerWon = true
				m.message = "Düşman donanmasını yok ettin!"
				return m, nil
			}
			m.message = fmt.Sprintf("Düşmanın %s gemisini batırdın!", ship.Name)
			m.waiting = true
			return m, botThink()
		}
	}
	return m, nil
}

func (m Model) botFire() (tea.Model, tea.Cmd) {
	c := m.bot.NextShot()
	result, ship := m.player.Fire(c)
	m.bot.Report(c, result)

	switch result {
	case game.FireMiss:
		m.message = fmt.Sprintf("Düşman %s'e ateş etti — ıska.", coordName(c))
	case game.FireHit:
		m.message = fmt.Sprintf("Düşman %s'te gemine isabet ettirdi!", coordName(c))
	case game.FireSunk:
		if m.player.AllSunk() {
			m.phase = phaseGameOver
			m.playerWon = false
			m.message = "Donanman yok edildi. Mağlup oldun."
			m.waiting = false
			return m, nil
		}
		m.message = fmt.Sprintf("Düşman %s gemini batırdı!", ship.Name)
	}

	m.waiting = false
	return m, nil
}

func (m Model) updateGameOver(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "r", "enter":
		return NewModel(), nil
	}
	return m, nil
}
