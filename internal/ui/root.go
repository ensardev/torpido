package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/lobby"
)

// Root is the top-level model for one connection. It shows the lobby, and swaps
// to a game when the player enters a room, swapping back when the game ends.
type Root struct {
	lobby    *lobby.Lobby
	name     string
	renderer *lipgloss.Renderer

	inGame bool
	lobbyM lobbyModel
	gameM  gameModel

	// the room/seat the player currently occupies, so we can free it on leave
	room *lobby.Room
	seat *lobby.Seat
}

// NewRoot returns the model a connection starts with: the lobby.
func NewRoot(l *lobby.Lobby, name string, renderer *lipgloss.Renderer) Root {
	return Root{
		lobby:    l,
		name:     name,
		renderer: renderer,
		lobbyM:   newLobbyModel(l, name, renderer),
	}
}

// enterRoomMsg tells the root to open the game screen for a room the player has
// just been seated in (as a bot game, a created room, or a joined match).
type enterRoomMsg struct {
	room *lobby.Room
	seat *lobby.Seat
}

// leaveGameMsg asks the root to end the current game and return to the lobby.
type leaveGameMsg struct{}

func leaveGame() tea.Msg { return leaveGameMsg{} }

func (m Root) Init() tea.Cmd { return m.lobbyM.Init() }

func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case enterRoomMsg:
		m.room, m.seat = msg.room, msg.seat
		m.gameM = newGameModel(msg.room, msg.seat, m.renderer)
		m.inGame = true
		return m, m.gameM.Init()

	case leaveGameMsg:
		if m.room != nil {
			m.lobby.Leave(m.room, m.seat)
			m.room, m.seat = nil, nil
		}
		m.inGame = false
		m.lobbyM = newLobbyModel(m.lobby, m.name, m.renderer)
		return m, m.lobbyM.Init()
	}

	if m.inGame {
		gm, cmd := m.gameM.Update(msg)
		m.gameM = gm.(gameModel)
		return m, cmd
	}
	lm, cmd := m.lobbyM.Update(msg)
	m.lobbyM = lm.(lobbyModel)
	return m, cmd
}

func (m Root) View() string {
	if m.inGame {
		return m.gameM.View()
	}
	return m.lobbyM.View()
}
