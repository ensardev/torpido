package ui

import "github.com/charmbracelet/lipgloss"

// styles holds every Lip Gloss style the UI uses. They are built per session
// from a renderer, so colors match the *connected* terminal. Over SSH each
// player has their own terminal with its own color support, so styles must come
// from that session's renderer (see server.MakeRenderer) rather than one shared
// global — otherwise the game renders with the server's colors, not yours.
type styles struct {
	water, ship, hit, miss, sunk      lipgloss.Style
	previewOK, previewBad, aim        lipgloss.Style
	logo, tag, dim, help, box         lipgloss.Style
	badgeYou, badgeFoe                lipgloss.Style
	rosterDone, rosterNow, rosterTodo lipgloss.Style
	win, lose                         lipgloss.Style
}

// newStyles builds the style set from a renderer. Colors are ANSI-256 codes.
func newStyles(r *lipgloss.Renderer) styles {
	c := func(code string) lipgloss.Color { return lipgloss.Color(code) }
	return styles{
		// Board cells: solid 2-wide blocks so ships and the sea look continuous.
		water: r.NewStyle().Background(c("17")).Foreground(c("25")),  // sea
		ship:  r.NewStyle().Background(c("22")),                      // your hull
		hit:   r.NewStyle().Background(c("160")).Foreground(c("231")), // struck
		miss:  r.NewStyle().Background(c("17")).Foreground(c("252")),  // splash on sea
		sunk:  r.NewStyle().Background(c("52")).Foreground(c("231")),  // sunk hull

		previewOK:  r.NewStyle().Background(c("40")),                     // fits here
		previewBad: r.NewStyle().Background(c("196")),                    // blocked
		aim:        r.NewStyle().Background(c("214")).Foreground(c("16")), // targeting reticle

		logo: r.NewStyle().Bold(true).Foreground(c("39")),
		tag:  r.NewStyle().Foreground(c("245")),
		dim:  r.NewStyle().Foreground(c("244")),
		help: r.NewStyle().Foreground(c("244")),
		box:  r.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(c("24")).Padding(0, 1),

		badgeYou: r.NewStyle().Background(c("28")).Foreground(c("231")).Bold(true).Padding(0, 1),
		badgeFoe: r.NewStyle().Background(c("130")).Foreground(c("231")).Bold(true).Padding(0, 1),

		rosterDone: r.NewStyle().Foreground(c("42")),
		rosterNow:  r.NewStyle().Background(c("39")).Foreground(c("16")).Bold(true).Padding(0, 1),
		rosterTodo: r.NewStyle().Foreground(c("240")),

		win:  r.NewStyle().Bold(true).Foreground(c("42")),
		lose: r.NewStyle().Bold(true).Foreground(c("196")),
	}
}
