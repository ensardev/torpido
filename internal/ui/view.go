package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/torpido/internal/game"
)

// This file holds the shared drawing helpers used by every screen. They take
// value-copy grids (never a live board), so rendering is always race-free.

type grid = [game.BoardSize][game.BoardSize]game.Cell

// coordName turns a coord into its player-facing name, e.g. {0,0} -> "A1".
func coordName(c game.Coord) string {
	return fmt.Sprintf("%c%d", 'A'+c.Col, c.Row+1)
}

// cellBlock renders one square as a 2-column colored block.
func (s styles) cellBlock(c game.Cell) string {
	switch c {
	case game.CellShip:
		return s.ship.Render("  ")
	case game.CellHit:
		return s.hit.Render("✖ ")
	case game.CellMiss:
		return s.miss.Render("○ ")
	case game.CellSunk:
		return s.sunk.Render("✖ ")
	default:
		return s.water.Render("· ")
	}
}

// renderBoard draws a 10x10 grid with A-J column and 1-10 row labels. Squares
// are drawn edge-to-edge (2 columns each) so ships and the sea look solid.
//
//   - aim, if set, highlights the targeting reticle (enemy board only).
//   - preview, if set, highlights where a ship is about to be placed;
//     previewValid tints it green (fits) or red (blocked).
func (s styles) renderBoard(g grid, aim *game.Coord, preview map[game.Coord]bool, previewValid bool) string {
	var sb strings.Builder

	sb.WriteString("   ")
	for c := 0; c < game.BoardSize; c++ {
		sb.WriteString(s.dim.Render(fmt.Sprintf("%-2s", string(rune('A'+c)))))
	}
	sb.WriteString("\n")

	for r := 0; r < game.BoardSize; r++ {
		sb.WriteString(s.dim.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < game.BoardSize; c++ {
			coord := game.Coord{Row: r, Col: c}
			switch {
			case preview != nil && preview[coord]:
				if previewValid {
					sb.WriteString(s.previewOK.Render("  "))
				} else {
					sb.WriteString(s.previewBad.Render("  "))
				}
			case aim != nil && *aim == coord:
				sb.WriteString(s.aim.Render("◎ "))
			default:
				sb.WriteString(s.cellBlock(g[r][c]))
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// boardPanel stacks a caption above a bordered board.
func (s styles) boardPanel(caption string, board string) string {
	return lipgloss.JoinVertical(lipgloss.Center, s.dim.Render(caption), s.box.Render(board))
}

// header is the logo line shown on every screen.
func (s styles) header() string {
	return s.logo.Render("🚢 TORPIDO") + "  " + s.tag.Render("terminal amiral battı")
}

// legend explains the glyphs, using the real colored blocks as a key.
func (s styles) legend() string {
	return s.ship.Render("  ") + s.dim.Render(" gemi   ") +
		s.hit.Render("✖ ") + s.dim.Render(" isabet   ") +
		s.miss.Render("○ ") + s.dim.Render(" ıska")
}

// screen wraps a screen body with the standard outer padding.
func screen(body string) string {
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
