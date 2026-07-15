package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/i18n"
	"github.com/ensardev/ssh-torpido/internal/lobby"
	"github.com/ensardev/ssh-torpido/internal/players"
)

// lobbyRefresh is how often the room list is refreshed so new rooms appear.
const lobbyRefresh = 2 * time.Second

type lobbyTickMsg time.Time

// lobbyNoticeMsg carries a transient message to show in the lobby.
type lobbyNoticeMsg string

// lobbyMode is whether the lobby is browsing rooms or typing an invite code.
type lobbyMode int

const (
	modeBrowse lobbyMode = iota
	modeJoinCode
	modeConfirmQuit
)

// lobbyModel is the screen a player sees after connecting: the list of joinable
// rooms and the actions to enter one.
type lobbyModel struct {
	lobby    *lobby.Lobby
	name     string
	fp       string
	store    *players.Store
	t        i18n.Strings
	renderer *lipgloss.Renderer
	styles   styles

	rooms  []lobby.RoomInfo
	cursor int
	notice string

	mode  lobbyMode
	input string
}

func newLobbyModel(l *lobby.Lobby, name, fp string, store *players.Store, t i18n.Strings, r *lipgloss.Renderer) lobbyModel {
	m := lobbyModel{
		lobby:    l,
		name:     name,
		fp:       fp,
		store:    store,
		t:        t,
		renderer: r,
		styles:   newStyles(r),
	}
	m.refresh()
	return m
}

func (m *lobbyModel) refresh() {
	m.rooms = m.lobby.PublicRooms()
	sort.SliceStable(m.rooms, func(i, j int) bool {
		a, b := m.rooms[i], m.rooms[j]
		if a.Kind != b.Kind {
			return a.Kind == lobby.BotRoom
		}
		if a.Kind == lobby.BotRoom {
			return a.Tier < b.Tier
		}
		return a.Code < b.Code
	})
	if m.cursor >= len(m.rooms) {
		m.cursor = len(m.rooms) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func lobbyTick() tea.Cmd {
	return tea.Tick(lobbyRefresh, func(t time.Time) tea.Msg { return lobbyTickMsg(t) })
}

func (m lobbyModel) Init() tea.Cmd { return lobbyTick() }

func (m lobbyModel) newSeat() *lobby.Seat {
	seat := lobby.NewHumanSeat(m.name)
	seat.Fingerprint = m.fp
	if m.fp != "" && m.store != nil {
		if rec, ok := m.store.Get(m.fp); ok {
			seat.Wins, seat.Losses = rec.Wins, rec.Losses
		}
	}
	return seat
}

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case lobbyTickMsg:
		m.refresh()
		return m, lobbyTick()
	case lobbyNoticeMsg:
		m.notice = string(msg)
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case modeJoinCode:
			return m.updateJoinCode(msg)
		case modeConfirmQuit:
			return m.updateConfirmQuit(msg)
		default:
			return m.updateBrowse(msg)
		}
	}
	return m, nil
}

func (m lobbyModel) updateConfirmQuit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "e", "enter", "ctrl+c":
		return m, tea.Quit
	default: // any other key cancels
		m.mode = modeBrowse
	}
	return m, nil
}

func (m lobbyModel) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// The lobby is a menu, so it navigates with the arrow keys and reserves the
	// letters for actions (unlike the game, which uses hjkl to move).
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		m.mode = modeConfirmQuit
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
		m.notice = ""
	case "down":
		if m.cursor < len(m.rooms)-1 {
			m.cursor++
		}
		m.notice = ""
	case "enter", " ":
		return m.selectRoom()
	case "c":
		return m.createRoom()
	case "h":
		return m.quickMatch()
	case "k":
		m.mode = modeJoinCode
		m.input = ""
		m.notice = ""
	}
	return m, nil
}

func (m lobbyModel) updateJoinCode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input = ""
	case "enter":
		code := m.input
		m.mode = modeBrowse
		m.input = ""
		return m.joinByCode(code)
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.String()) == 1 && len(m.input) < 4 {
			m.input += strings.ToUpper(msg.String())
		}
	}
	return m, nil
}

func (m lobbyModel) selectRoom() (tea.Model, tea.Cmd) {
	if len(m.rooms) == 0 {
		return m, nil
	}
	info := m.rooms[m.cursor]
	if info.HasPassword {
		m.notice = m.t.LPasswordSoon
		return m, nil
	}
	return m.joinByCode(info.Code)
}

func (m lobbyModel) joinByCode(code string) (tea.Model, tea.Cmd) {
	l, seat, t := m.lobby, m.newSeat(), m.t
	return m, func() tea.Msg {
		room, err := l.JoinByCode(code, seat, "")
		if err != nil {
			return lobbyNoticeMsg(joinErrText(err, t))
		}
		return enterRoomMsg{room: room, seat: seat}
	}
}

// joinErrText maps a lobby join error to a translated message.
func joinErrText(err error, t i18n.Strings) string {
	switch {
	case errors.Is(err, lobby.ErrNoRoom):
		return t.LErrNoRoom
	case errors.Is(err, lobby.ErrRoomFull):
		return t.LErrRoomFull
	default:
		return err.Error()
	}
}

func (m lobbyModel) createRoom() (tea.Model, tea.Cmd) {
	l, seat := m.lobby, m.newSeat()
	return m, func() tea.Msg {
		room := l.CreateRoom(seat, "", true) // private: joinable only by its code
		return enterRoomMsg{room: room, seat: seat}
	}
}

func (m lobbyModel) quickMatch() (tea.Model, tea.Cmd) {
	l, seat := m.lobby, m.newSeat()
	return m, func() tea.Msg {
		room := l.QuickMatch(seat)
		return enterRoomMsg{room: room, seat: seat}
	}
}

// tierStyle picks the accent color for a bot difficulty.
func (s styles) tierStyle(d game.Difficulty) lipgloss.Style {
	switch d {
	case game.Rookie:
		return s.tierRookie
	case game.Admiral:
		return s.tierAdmiral
	case game.SeaWolf:
		return s.tierWolf
	default:
		return s.dim
	}
}

func (m lobbyModel) View() string {
	s := m.styles
	var rows []string
	for i, info := range m.rooms {
		var line string
		if info.Kind == lobby.BotRoom {
			line = fmt.Sprintf("%s %-12s %s",
				s.tierStyle(info.Tier).Render("●"),
				info.HostName,
				s.dim.Render(m.t.LBotWaiting))
		} else {
			lock := ""
			if info.HasPassword {
				lock = "🔒 "
			}
			host := info.HostName
			if host == "" {
				host = m.t.LPlayer
			}
			record := ""
			if info.HostWins+info.HostLosses > 0 {
				record = s.tag.Render(fmt.Sprintf(" (%d/%d)", info.HostWins, info.HostLosses))
			}
			line = fmt.Sprintf("%s⚔ %s %s%s",
				lock, s.logo.Render(info.Code),
				s.dim.Render(fmt.Sprintf(m.t.LHumanWaitingFmt, host, info.Players)), record)
		}
		if i == m.cursor {
			line = s.rosterNow.Render("▸ " + strings.TrimPrefix(line, " "))
		} else {
			line = "  " + line
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = append(rows, s.dim.Render("  "+m.t.LNoRooms))
	}
	list := s.box.Render(strings.Join(rows, "\n"))

	footer := s.help.Render(m.t.LFooter)
	switch m.mode {
	case modeJoinCode:
		footer = s.badgeYou.Render(m.t.LCode+m.input+"_") + "  " + s.help.Render(m.t.LCodeHelp)
	case modeConfirmQuit:
		footer = s.badgeFoe.Render(m.t.LQuitConfirm)
	}

	notice := ""
	if m.notice != "" {
		notice = s.tag.Render(m.notice) + "\n"
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(m.t.Tagline),
		"",
		s.dim.Render(m.t.LOpenRooms),
		list,
		"",
		notice+footer,
	)
	return screen(body)
}
