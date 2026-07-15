package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/i18n"
	"github.com/ensardev/ssh-torpido/internal/lobby"
	"github.com/ensardev/ssh-torpido/internal/players"
)

type rootScreen int

const (
	rootWelcome rootScreen = iota
	rootLobby
	rootGame
)

// Root is the top-level model for one connection. It walks the player from the
// welcome screen to the lobby to a game, and carries the chosen language.
type Root struct {
	lobby    *lobby.Lobby
	name     string
	fp       string
	store    *players.Store
	renderer *lipgloss.Renderer

	lang   i18n.Lang
	t      i18n.Strings
	screen rootScreen

	welcomeM welcomeModel
	lobbyM   lobbyModel
	gameM    gameModel

	// the room/seat the player currently occupies, so we can free it on leave
	room *lobby.Room
	seat *lobby.Seat
}

// NewRoot returns the model a connection starts with: the welcome screen. fp is
// the player's SSH-key fingerprint (empty for keyless guests) and store is the
// shared persistent player record.
func NewRoot(l *lobby.Lobby, name, fp string, store *players.Store, renderer *lipgloss.Renderer) Root {
	return Root{
		lobby:    l,
		name:     name,
		fp:       fp,
		store:    store,
		renderer: renderer,
		lang:     i18n.EN,
		t:        i18n.For(i18n.EN),
		screen:   rootWelcome,
		welcomeM: newWelcomeModel(i18n.EN, renderer),
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

func (m Root) Init() tea.Cmd { return m.welcomeM.Init() }

func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startLobbyMsg:
		m.lang = msg.lang
		m.t = i18n.For(msg.lang)
		m.lobbyM = newLobbyModel(m.lobby, m.name, m.fp, m.store, m.t, m.renderer)
		m.screen = rootLobby
		return m, m.lobbyM.Init()

	case enterRoomMsg:
		m.room, m.seat = msg.room, msg.seat
		m.gameM = newGameModel(msg.room, msg.seat, m.fp, m.store, m.t, m.renderer)
		m.screen = rootGame
		return m, m.gameM.Init()

	case leaveGameMsg:
		if m.room != nil {
			m.lobby.Leave(m.room, m.seat)
			m.room, m.seat = nil, nil
		}
		m.lobbyM = newLobbyModel(m.lobby, m.name, m.fp, m.store, m.t, m.renderer)
		m.screen = rootLobby
		return m, m.lobbyM.Init()
	}

	switch m.screen {
	case rootWelcome:
		wm, cmd := m.welcomeM.Update(msg)
		m.welcomeM = wm.(welcomeModel)
		return m, cmd
	case rootLobby:
		lm, cmd := m.lobbyM.Update(msg)
		m.lobbyM = lm.(lobbyModel)
		return m, cmd
	default:
		gm, cmd := m.gameM.Update(msg)
		m.gameM = gm.(gameModel)
		return m, cmd
	}
}

func (m Root) View() string {
	switch m.screen {
	case rootWelcome:
		return m.welcomeM.View()
	case rootLobby:
		return m.lobbyM.View()
	default:
		return m.gameM.View()
	}
}
