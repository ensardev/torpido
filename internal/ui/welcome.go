package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/i18n"
	"github.com/ensardev/ssh-torpido/internal/players"
)

// welcomeAnimTick is how fast the torpedo animation advances.
const welcomeAnimTick = 110 * time.Millisecond

// blockLetters is a small 5-row block font for the TORPIDO wordmark.
var blockLetters = map[rune][5]string{
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
	'I': {"█", "█", "█", "█", "█"},
	'D': {"████ ", "█   █", "█   █", "█   █", "████ "},
}

// banner builds the 5 rows of a word from the block font.
func banner(word string) []string {
	rows := make([]string, 5)
	for i := 0; i < 5; i++ {
		var parts []string
		for _, ch := range word {
			parts = append(parts, blockLetters[ch][i])
		}
		rows[i] = strings.Join(parts, " ")
	}
	return rows
}

// bannerShades colors the wordmark rows top-to-bottom for a bit of depth.
var bannerShades = []string{"51", "45", "39", "33", "27"}

type welcomePage int

const (
	pageMenu welcomePage = iota
	pageHowTo
	pageWhatSSH
	pageAbout
	pageLeaderboard
	pageNickname
)

// welcome menu item indices.
const (
	miPlay = iota
	miLeaderboard
	miNickname
	miHowTo
	miWhatSSH
	miAbout
	miLanguage
	miQuit
	miCount
)

type welcomeTickMsg struct{}

// startLobbyMsg tells the root to leave the welcome screen for the lobby, in the
// language the player picked.
type startLobbyMsg struct{ lang i18n.Lang }

// setNickMsg tells the root the player picked a new nickname.
type setNickMsg struct{ nick string }

type welcomeModel struct {
	lang     i18n.Lang
	t        i18n.Strings
	nick     string
	fp       string
	store    *players.Store
	renderer *lipgloss.Renderer
	styles   styles

	page        welcomePage
	cursor      int
	frame       int
	width       int
	input       string
	notice      string
	confirmQuit bool
}

func newWelcomeModel(lang i18n.Lang, nick, fp string, store *players.Store, r *lipgloss.Renderer) welcomeModel {
	return welcomeModel{
		lang:     lang,
		t:        i18n.For(lang),
		nick:     nick,
		fp:       fp,
		store:    store,
		renderer: r,
		styles:   newStyles(r),
		width:    80,
	}
}

func welcomeTick() tea.Cmd {
	return tea.Tick(welcomeAnimTick, func(time.Time) tea.Msg { return welcomeTickMsg{} })
}

func (m welcomeModel) Init() tea.Cmd { return welcomeTick() }

func (m welcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case welcomeTickMsg:
		m.frame++
		return m, welcomeTick()
	case tea.KeyMsg:
		switch m.page {
		case pageMenu:
			return m.updateMenu(msg)
		case pageNickname:
			return m.updateNickname(msg)
		default: // info & leaderboard pages: any key returns to the menu
			m.page = pageMenu
			m.notice = ""
			return m, nil
		}
	}
	return m, nil
}

func (m welcomeModel) updateNickname(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.page = pageMenu
		m.input, m.notice = "", ""
	case "enter":
		nick := strings.TrimSpace(m.input)
		if nick == "" || m.store == nil || m.fp == "" {
			return m, nil
		}
		if m.store.SetNick(m.fp, nick) {
			m.nick, m.input, m.notice = nick, "", m.t.WNickSet
			return m, func() tea.Msg { return setNickMsg{nick: nick} }
		}
		m.notice = m.t.WNickTaken
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.String()) == 1 && len(m.input) < 16 {
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m welcomeModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmQuit {
		switch msg.String() {
		case "y", "e", "enter", "ctrl+c":
			return m, tea.Quit
		default: // any other key cancels
			m.confirmQuit = false
		}
		return m, nil
	}
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
		m.confirmQuit = true
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < miCount-1 {
			m.cursor++
		}
	case "left", "right", "h", "l":
		if m.cursor == miLanguage {
			m.setLang(m.lang.Next())
		}
	case "enter", " ":
		return m.selectItem()
	}
	return m, nil
}

func (m *welcomeModel) setLang(l i18n.Lang) {
	m.lang = l
	m.t = i18n.For(l)
}

func (m welcomeModel) selectItem() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case miPlay:
		return m, func() tea.Msg { return startLobbyMsg{lang: m.lang} }
	case miLeaderboard:
		m.page = pageLeaderboard
	case miNickname:
		m.page, m.input, m.notice = pageNickname, "", ""
	case miHowTo:
		m.page = pageHowTo
	case miWhatSSH:
		m.page = pageWhatSSH
	case miAbout:
		m.page = pageAbout
	case miLanguage:
		m.setLang(m.lang.Next())
	case miQuit:
		return m, tea.Quit
	}
	return m, nil
}
