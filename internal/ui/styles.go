package ui

import "github.com/charmbracelet/lipgloss"

// Each board square is drawn as a 2-column block with NO gap between squares,
// so same-colored neighbours merge into one solid shape. That is what makes a
// ship look like one continuous hull both horizontally and vertically, and what
// turns the empty squares into one continuous sea.

// Cell block styles. Colors are ANSI-256 codes so they work in any terminal.
var (
	styleWater = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("25"))  // sea
	styleShip  = lipgloss.NewStyle().Background(lipgloss.Color("22"))                                    // your hull
	styleHit   = lipgloss.NewStyle().Background(lipgloss.Color("160")).Foreground(lipgloss.Color("231")) // struck
	styleMiss  = lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("252"))  // splash on sea
	styleSunk  = lipgloss.NewStyle().Background(lipgloss.Color("52")).Foreground(lipgloss.Color("231"))  // sunk hull

	stylePreviewOK  = lipgloss.NewStyle().Background(lipgloss.Color("40"))                                     // fits here
	stylePreviewBad = lipgloss.NewStyle().Background(lipgloss.Color("196"))                                   // blocked
	styleAim        = lipgloss.NewStyle().Background(lipgloss.Color("214")).Foreground(lipgloss.Color("16"))  // targeting reticle
)

// Chrome around the boards.
var (
	styleLogo = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleTag  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleDim  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleHelp = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleBox  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("24")).Padding(0, 1)

	styleBadgeYou = lipgloss.NewStyle().Background(lipgloss.Color("28")).Foreground(lipgloss.Color("231")).Bold(true).Padding(0, 1)
	styleBadgeFoe = lipgloss.NewStyle().Background(lipgloss.Color("130")).Foreground(lipgloss.Color("231")).Bold(true).Padding(0, 1)

	styleRosterDone = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleRosterNow  = lipgloss.NewStyle().Background(lipgloss.Color("39")).Foreground(lipgloss.Color("16")).Bold(true).Padding(0, 1)
	styleRosterTodo = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	styleWin  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	styleLose = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
)
