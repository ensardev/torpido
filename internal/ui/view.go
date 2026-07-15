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

// cellBlock renders one square as a 2-column colored block.
func cellBlock(c game.Cell) string {
	switch c {
	case game.CellShip:
		return styleShip.Render("  ")
	case game.CellHit:
		return styleHit.Render("✖ ")
	case game.CellMiss:
		return styleMiss.Render("○ ")
	case game.CellSunk:
		return styleSunk.Render("✖ ")
	default:
		return styleWater.Render("· ")
	}
}

// renderBoard draws a 10x10 grid with A-J column and 1-10 row labels. Squares
// are drawn edge-to-edge (2 columns each) so ships and the sea look solid.
//
//   - reveal shows un-hit ships (use for your own board, not the enemy's).
//   - aim, if set, highlights the targeting reticle (enemy board only).
//   - preview, if set, highlights where a ship is about to be placed;
//     previewValid tints it green (fits) or red (blocked).
func renderBoard(b *game.Board, reveal bool, aim *game.Coord, preview map[game.Coord]bool, previewValid bool) string {
	var sb strings.Builder

	// Column header: A B C ... aligned to the 2-wide squares.
	sb.WriteString("   ")
	for c := 0; c < game.BoardSize; c++ {
		sb.WriteString(styleDim.Render(fmt.Sprintf("%-2s", string(rune('A'+c)))))
	}
	sb.WriteString("\n")

	for r := 0; r < game.BoardSize; r++ {
		sb.WriteString(styleDim.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < game.BoardSize; c++ {
			coord := game.Coord{Row: r, Col: c}
			switch {
			case preview != nil && preview[coord]:
				if previewValid {
					sb.WriteString(stylePreviewOK.Render("  "))
				} else {
					sb.WriteString(stylePreviewBad.Render("  "))
				}
			case aim != nil && *aim == coord:
				sb.WriteString(styleAim.Render("◎ "))
			default:
				sb.WriteString(cellBlock(b.StateAt(coord, reveal)))
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// boardPanel stacks a caption above a bordered board.
func boardPanel(caption, board string) string {
	return lipgloss.JoinVertical(lipgloss.Center, styleDim.Render(caption), styleBox.Render(board))
}

// header is the logo line shown on every screen.
func header() string {
	return styleLogo.Render("🚢 TORPIDO") + "  " + styleTag.Render("terminal amiral battı")
}

// legend explains the glyphs, using the real colored blocks as a key.
func legend() string {
	return styleShip.Render("  ") + styleDim.Render(" gemi   ") +
		styleHit.Render("✖ ") + styleDim.Render(" isabet   ") +
		styleMiss.Render("○ ") + styleDim.Render(" ıska")
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

	// Roster of ships: done ones ticked, current one highlighted, rest dimmed.
	var roster []string
	for i, s := range m.fleet {
		label := fmt.Sprintf("%s (%d)", s.Name, s.Size)
		switch {
		case i < m.placeIndex:
			roster = append(roster, styleRosterDone.Render("✔ "+label))
		case i == m.placeIndex:
			roster = append(roster, styleRosterNow.Render(label))
		default:
			roster = append(roster, styleRosterTodo.Render(label))
		}
	}

	help := styleHelp.Render(fmt.Sprintf(
		"ok tuşları/hjkl taşı · r döndür (%s) · enter yerleştir · q çık", orient))

	body := lipgloss.JoinVertical(lipgloss.Left,
		header(),
		"",
		styleDim.Render("Donanmanı yerleştir:"),
		strings.Join(roster, "   "),
		"",
		boardPanel("SENİN SULARIN", renderBoard(m.player, true, nil, preview, valid)),
		"",
		legend(),
		"",
		help,
	)
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

	turn := styleBadgeYou.Render("SIRA: SEN")
	if m.waiting {
		turn = styleBadgeFoe.Render("DÜŞMAN NİŞAN ALIYOR…")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		header(),
		"",
		turn+"   "+m.message,
		"",
		boards,
		"",
		legend(),
		"",
		styleHelp.Render("ok tuşları/hjkl nişan al · enter ateş · q çık"),
	)
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

	body := lipgloss.JoinVertical(lipgloss.Left,
		header(),
		"",
		banner+"   "+m.message,
		"",
		boards,
		"",
		styleHelp.Render("r/enter tekrar oyna · q çık"),
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
