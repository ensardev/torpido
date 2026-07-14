package ui

import "github.com/charmbracelet/lipgloss"

// Glyphs drawn in each square.
const (
	glyphWater = "·"
	glyphShip  = "█"
	glyphHit   = "✖"
	glyphMiss  = "○"
	glyphSunk  = "▓"
)

// Lip Gloss styles. Colors are ANSI-256 codes so they work in any terminal.
var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleDim   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleHelp  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	styleWater = lipgloss.NewStyle().Foreground(lipgloss.Color("24"))
	styleShip  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleHit   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	styleMiss  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleSunk  = lipgloss.NewStyle().Foreground(lipgloss.Color("88")).Bold(true)

	stylePreviewOK  = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("42"))
	stylePreviewBad = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("196"))
	styleAim        = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("208")).Bold(true)

	styleBox  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("24")).Padding(0, 1)
	styleWin  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	styleLose = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
)
