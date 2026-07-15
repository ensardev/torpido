package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m welcomeModel) View() string {
	switch m.page {
	case pageHowTo:
		return m.viewInfo(m.t.WHowToTitle, m.t.WHowToBody)
	case pageWhatSSH:
		return m.viewInfo(m.t.WWhatSSHTitle, m.t.WWhatSSHBody)
	case pageAbout:
		return m.viewInfo(m.t.WAboutTitle, m.t.WAboutBody)
	case pageLeaderboard:
		return m.viewLeaderboard()
	case pageNickname:
		return m.viewNickname()
	default:
		return m.viewMenu()
	}
}

// renderBanner colors the wordmark rows with a top-to-bottom gradient.
func (m welcomeModel) renderBanner() string {
	rows := banner("TORPIDO")
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(bannerShades[i])).Render(row)
	}
	return strings.Join(out, "\n")
}

// renderTorpedo draws a torpedo sailing across a waterline, animated by frame.
func (m welcomeModel) renderTorpedo(width int) string {
	if width < 8 {
		width = 8
	}
	sprite := []rune("══►") // runes, not bytes — each glyph is one cell
	line := []rune(strings.Repeat("·", width))
	start := m.frame%(width+len(sprite)) - len(sprite) + 1
	for i, ch := range sprite {
		if p := start + i; p >= 0 && p < width {
			line[p] = ch
		}
	}
	// Dim water dots with a bright torpedo where it currently sits.
	var b strings.Builder
	for i, ch := range line {
		if i >= start && i < start+len(sprite) {
			b.WriteString(m.styles.tierAdmiral.Render(string(ch)))
		} else {
			b.WriteString(m.styles.dim.Render(string(ch)))
		}
	}
	return b.String()
}

func (m welcomeModel) viewMenu() string {
	s := m.styles
	logo := m.renderBanner()
	bannerW := lipgloss.Width(logo)

	tagline := s.tag.Render(m.t.Tagline)
	torpedo := m.renderTorpedo(bannerW)

	items := []string{
		m.t.WPlay,
		m.t.WLeaderboard,
		fmt.Sprintf("%s: %s", m.t.WNickname, m.nick),
		m.t.WHowTo,
		m.t.WWhatSSH,
		m.t.WAbout,
		fmt.Sprintf("%s: %s", m.t.WLanguage, m.lang.Label()),
		m.t.WQuit,
	}
	var menu []string
	for i, it := range items {
		if i == m.cursor {
			menu = append(menu, s.logo.Render("▸ ")+s.rosterNow.Render(it))
		} else {
			menu = append(menu, "  "+s.dim.Render(it))
		}
	}

	nav := s.help.Render(m.t.WNav)
	if m.confirmQuit {
		nav = s.badgeFoe.Render(m.t.LQuitConfirm)
	}

	block := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		tagline,
		"",
		torpedo,
		"",
		"",
		lipgloss.JoinVertical(lipgloss.Left, menu...),
		"",
		nav,
		s.dim.Render("ssh torpido.dev"),
		s.tag.Render("by ")+s.logo.Render("ensar.dev"),
	)
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Padding(1, 0).Render(block)
}

func (m welcomeModel) viewLeaderboard() string {
	s := m.styles
	var rows []string
	if m.store != nil {
		for i, r := range m.store.Top(10) {
			marker, nickStyle := "   ", s.dim
			if r.Fingerprint == m.fp && m.fp != "" {
				marker, nickStyle = s.logo.Render(" ▸ "), s.rosterDone
			}
			rows = append(rows, fmt.Sprintf("%s%2d.  %s  %s",
				marker, i+1, nickStyle.Render(fmt.Sprintf("%-16s", r.Nick)), s.wl(r.Wins, r.Losses)))
		}
	}
	if len(rows) == 0 {
		rows = append(rows, s.dim.Render(m.t.WLbEmpty))
	}

	yourRank := ""
	if m.store != nil && m.fp != "" {
		if rec, ok := m.store.Get(m.fp); ok {
			yourRank = s.tag.Render(fmt.Sprintf(m.t.WLbYouRankFmt, m.store.Rank(m.fp))) + s.wl(rec.Wins, rec.Losses)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		s.logo.Render(m.t.WLbTitle),
		"",
		strings.Join(rows, "\n"),
		"",
		yourRank,
		"",
		s.help.Render(m.t.WInfoBack),
	)
	return s.framed(m.width, content)
}

func (m welcomeModel) viewNickname() string {
	s := m.styles
	notice := ""
	if m.notice != "" {
		notice = s.tag.Render(m.notice) + "\n\n"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		s.logo.Render(m.t.WNickTitle),
		"",
		s.badgeYou.Render(" "+m.input+"_ "),
		"",
		notice+s.help.Render(m.t.WNickHelp),
	)
	return s.framed(m.width, content)
}

func (m welcomeModel) viewInfo(title, body string) string {
	s := m.styles
	content := lipgloss.JoinVertical(lipgloss.Left,
		s.logo.Render(title),
		"",
		body,
		"",
		s.help.Render(m.t.WInfoBack),
	)
	return s.framed(m.width, content)
}
