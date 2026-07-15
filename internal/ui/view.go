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
//   - reveal shows un-hit ships (use for your own board, not the enemy's).
//   - aim, if set, highlights the targeting reticle (enemy board only).
//   - preview, if set, highlights where a ship is about to be placed;
//     previewValid tints it green (fits) or red (blocked).
func (s styles) renderBoard(b *game.Board, reveal bool, aim *game.Coord, preview map[game.Coord]bool, previewValid bool) string {
	var sb strings.Builder

	// Column header: A B C ... aligned to the 2-wide squares.
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
				sb.WriteString(s.cellBlock(b.StateAt(coord, reveal)))
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// boardPanel stacks a caption above a bordered board.
func (s styles) boardPanel(caption, board string) string {
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

func (m Model) viewPlacement() string {
	s := m.styles
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
	for i, sh := range m.fleet {
		label := fmt.Sprintf("%s (%d)", sh.Name, sh.Size)
		switch {
		case i < m.placeIndex:
			roster = append(roster, s.rosterDone.Render("✔ "+label))
		case i == m.placeIndex:
			roster = append(roster, s.rosterNow.Render(label))
		default:
			roster = append(roster, s.rosterTodo.Render(label))
		}
	}

	help := s.help.Render(fmt.Sprintf(
		"ok tuşları/hjkl taşı · r döndür (%s) · enter yerleştir · q çık", orient))

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		s.dim.Render("Donanmanı yerleştir:"),
		strings.Join(roster, "   "),
		"",
		s.boardPanel("SENİN SULARIN", s.renderBoard(m.player, true, nil, preview, valid)),
		"",
		s.legend(),
		"",
		help,
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

func (m Model) viewBattle() string {
	s := m.styles
	var aim *game.Coord
	if !m.waiting {
		aim = &m.aim
	}
	own := s.boardPanel("SENİN SULARIN", s.renderBoard(m.player, true, nil, nil, false))
	enemy := s.boardPanel("DÜŞMAN SULARI", s.renderBoard(m.enemy, false, aim, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	turn := s.badgeYou.Render("SIRA: SEN")
	if m.waiting {
		turn = s.badgeFoe.Render("DÜŞMAN NİŞAN ALIYOR…")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		turn+"   "+m.message,
		"",
		boards,
		"",
		s.legend(),
		"",
		s.help.Render("ok tuşları/hjkl nişan al · enter ateş · q çık"),
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

func (m Model) viewGameOver() string {
	s := m.styles
	banner := s.win.Render("★  ZAFER  ★")
	if !m.playerWon {
		banner = s.lose.Render("✖  MAĞLUBİYET  ✖")
	}
	own := s.boardPanel("SENİN SULARIN", s.renderBoard(m.player, true, nil, nil, false))
	enemy := s.boardPanel("DÜŞMAN SULARI", s.renderBoard(m.enemy, true, nil, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		banner+"   "+m.message,
		"",
		boards,
		"",
		s.help.Render("r/enter tekrar oyna · q çık"),
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
