package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/torpido/internal/game"
)

func (m gameModel) View() string {
	switch m.phase {
	case gameWaiting:
		return m.viewWaiting()
	case gamePlacing:
		return m.viewPlacing()
	case gamePlaceWait:
		return m.viewPlaceWait()
	case gameBattle:
		return m.viewBattle()
	case gameOver:
		return m.viewOver()
	}
	return ""
}

// opponentName is who you're up against, for headers.
func (m gameModel) opponentName() string {
	if m.snap.OppName != "" {
		return m.snap.OppName
	}
	return "rakip"
}

func (m gameModel) viewWaiting() string {
	s := m.styles
	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		s.badgeYou.Render("ODA: "+m.room.Code),
		"",
		s.tag.Render("Rakip bekleniyor…"),
		s.dim.Render("Bu kodu arkadaşına gönder — o da ‘kodla katıl’ ile girsin."),
		"",
		s.help.Render("q lobiye dön"),
	)
	return screen(body)
}

func (m gameModel) viewPlacing() string {
	s := m.styles
	st := m.fleet[m.placeIndex]
	coords := game.ShipCoords(m.cursor, st.Size, m.orientation)
	valid := m.previewValid(coords)
	preview := make(map[game.Coord]bool, len(coords))
	for _, c := range coords {
		preview[c] = true
	}

	orient := "yatay"
	if m.orientation == game.Vertical {
		orient = "dikey"
	}

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

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header()+"  "+s.tag.Render("· "+m.opponentName()+"'e karşı"),
		"",
		s.dim.Render("Donanmanı yerleştir:"),
		strings.Join(roster, "   "),
		"",
		s.boardPanel("SENİN SULARIN", s.renderBoard(m.snap.You, nil, preview, valid)),
		"",
		s.legend(),
		"",
		s.help.Render(fmt.Sprintf("ok tuşları/hjkl taşı · r döndür (%s) · enter yerleştir · q çık", orient)),
	)
	return screen(body)
}

func (m gameModel) viewPlaceWait() string {
	s := m.styles
	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		s.badgeYou.Render("HAZIRSIN"),
		"",
		s.boardPanel("SENİN SULARIN", s.renderBoard(m.snap.You, nil, nil, false)),
		"",
		s.tag.Render(m.opponentName()+" hâlâ donanmasını yerleştiriyor…"),
		"",
		s.help.Render("q lobiye dön"),
	)
	return screen(body)
}

func (m gameModel) viewBattle() string {
	s := m.styles
	var aim *game.Coord
	if m.snap.YourTurn {
		aim = &m.aim
	}
	own := s.boardPanel("SENİN SULARIN", s.renderBoard(m.snap.You, nil, nil, false))
	enemy := s.boardPanel("DÜŞMAN SULARI", s.renderBoard(m.snap.Enemy, aim, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	var turn string
	if m.snap.YourTurn {
		turn = s.badgeYou.Render("SIRA: SEN")
	} else {
		turn = s.badgeFoe.Render(strings.ToUpper(m.opponentName()) + " NİŞAN ALIYOR…")
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header()+"  "+s.tag.Render("· "+m.opponentName()+"'e karşı"),
		"",
		turn+"   "+m.message,
		"",
		boards,
		"",
		s.legend(),
		"",
		s.help.Render("ok tuşları/hjkl nişan al · enter ateş · q lobiye dön"),
	)
	return screen(body)
}

func (m gameModel) viewOver() string {
	s := m.styles
	banner := s.win.Render("★  ZAFER  ★")
	msg := m.opponentName() + " donanmasını yok ettin!"
	if !m.snap.YouWon {
		banner = s.lose.Render("✖  MAĞLUBİYET  ✖")
		msg = m.opponentName() + " seni yendi."
	}
	own := s.boardPanel("SENİN SULARIN", s.renderBoard(m.snap.You, nil, nil, false))
	enemy := s.boardPanel("DÜŞMAN SULARI", s.renderBoard(m.snap.EnemyFull, nil, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	body := lipgloss.JoinVertical(lipgloss.Left,
		banner+"   "+msg,
		"",
		boards,
		"",
		s.help.Render("enter/q lobiye dön"),
	)
	return screen(body)
}
