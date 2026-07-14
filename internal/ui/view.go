package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/torpido/internal/game"
)

// View renders the current screen. It is the single entry point Bubble Tea
// calls; it just dispatches to the right screen for the current phase.
func (m Model) View() string {
	switch m.phase {
	case phasePlacement:
		return m.viewPlacement()
	case phaseBattle:
		return m.viewBattle()
	case phaseGameOver:
		return m.viewGameOver()
	}
	return ""
}

// coordName turns a coord into its player-facing name, e.g. {0,0} -> "A1".
func coordName(c game.Coord) string {
	return fmt.Sprintf("%c%d", 'A'+c.Col, c.Row+1)
}

// cellGlyph maps a cell state to its glyph and style.
func cellGlyph(c game.Cell) (string, lipgloss.Style) {
	switch c {
	case game.CellShip:
		return glyphShip, styleShip
	case game.CellHit:
		return glyphHit, styleHit
	case game.CellMiss:
		return glyphMiss, styleMiss
	case game.CellSunk:
		return glyphSunk, styleSunk
	default:
		return glyphWater, styleWater
	}
}

// renderBoard draws a 10x10 grid with A-J column and 1-10 row labels.
//
//   - reveal shows un-hit ships (use for your own board, not the enemy's).
//   - aim, if set, highlights the targeting cursor (enemy board only).
//   - preview, if set, highlights where a ship is about to be placed;
//     previewValid tints it green (fits) or red (blocked).
func renderBoard(b *game.Board, reveal bool, aim *game.Coord, preview map[game.Coord]bool, previewValid bool) string {
	var sb strings.Builder

	// Column header: A B C ...
	sb.WriteString("   ")
	for c := 0; c < game.BoardSize; c++ {
		sb.WriteString(styleDim.Render(string(rune('A'+c)) + " "))
	}
	sb.WriteString("\n")

	for r := 0; r < game.BoardSize; r++ {
		sb.WriteString(styleDim.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < game.BoardSize; c++ {
			coord := game.Coord{Row: r, Col: c}
			glyph, style := cellGlyph(b.StateAt(coord, reveal))

			if preview != nil && preview[coord] {
				glyph = glyphShip
				if previewValid {
					style = stylePreviewOK
				} else {
					style = stylePreviewBad
				}
			}
			if aim != nil && *aim == coord {
				style = styleAim
			}

			sb.WriteString(style.Render(glyph) + " ")
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// boardPanel stacks a caption above a bordered board.
func boardPanel(caption, board string) string {
	return lipgloss.JoinVertical(lipgloss.Center, styleDim.Render(caption), styleBox.Render(board))
}

func (m Model) viewPlacement() string {
	st := m.fleet[m.placeIndex]
	coords := game.ShipCoords(m.cursor, st.Size, m.orientation)
	valid := m.player.CanPlace(coords)
	preview := make(map[game.Coord]bool, len(coords))
	for _, c := range coords {
		preview[c] = true
	}

	orient := "yatay"
	if m.orientation == game.Vertical {
		orient = "dikey"
	}

	title := styleTitle.Render("torpido — donanmanı yerleştir")
	info := fmt.Sprintf("Yerleştiriliyor: %s (%d kare)   —   %d/%d gemi",
		st.Name, st.Size, m.placeIndex+1, len(m.fleet))
	help := styleHelp.Render(fmt.Sprintf(
		"ok tuşları/hjkl taşı · r döndür (%s) · enter yerleştir · q çık", orient))

	body := lipgloss.JoinVertical(lipgloss.Left,
		title, "", info, "", boardPanel("SENİN SULARIN", renderBoard(m.player, true, nil, preview, valid)), "", help)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

func (m Model) viewBattle() string {
	var aim *game.Coord
	if !m.waiting {
		aim = &m.aim
	}
	own := boardPanel("SENİN SULARIN", renderBoard(m.player, true, nil, nil, false))
	enemy := boardPanel("DÜŞMAN SULARI", renderBoard(m.enemy, false, aim, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	status := m.message
	if m.waiting {
		status = m.message + "  " + styleDim.Render("(düşman nişan alıyor…)")
	}
	help := styleHelp.Render("ok tuşları/hjkl nişan al · enter ateş · q çık")

	body := lipgloss.JoinVertical(lipgloss.Left,
		styleTitle.Render("torpido"), "", boards, "", status, "", help)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

func (m Model) viewGameOver() string {
	banner := styleWin.Render("★  ZAFER  ★")
	if !m.playerWon {
		banner = styleLose.Render("✖  MAĞLUBİYET  ✖")
	}
	own := boardPanel("SENİN SULARIN", renderBoard(m.player, true, nil, nil, false))
	enemy := boardPanel("DÜŞMAN SULARI", renderBoard(m.enemy, true, nil, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)
	help := styleHelp.Render("r/enter tekrar oyna · q çık")

	body := lipgloss.JoinVertical(lipgloss.Left,
		banner, "", m.message, "", boards, "", help)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
