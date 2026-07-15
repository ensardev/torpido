package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/i18n"
	"github.com/ensardev/ssh-torpido/internal/lobby"
)

func placeFleet(room *lobby.Room, side game.Side) {
	ships := []struct {
		name string
		size int
	}{{"Carrier", 5}, {"Battleship", 4}, {"Cruiser", 3}, {"Submarine", 3}, {"Destroyer", 2}}
	for i, sh := range ships {
		coords := game.ShipCoords(game.Coord{Row: i, Col: 0}, sh.size, game.Horizontal)
		room.PlaceShip(side, sh.name, sh.size, coords)
	}
	room.FinishPlacing(side)
}

// TestTwoPlayersTakeTurns drives two game models sharing one room through the
// core loop: SideA fires, and after SideB refreshes (as its update channel would
// make it), the turn has passed and the shot shows on SideB's board.
func TestTwoPlayersTakeTurns(t *testing.T) {
	l := lobby.New(nil)
	a := lobby.NewHumanSeat("Ali")
	room := l.CreateRoom(a, "", false)
	b := lobby.NewHumanSeat("Veli")
	if _, err := l.JoinByCode(room.Code, b, ""); err != nil {
		t.Fatalf("Veli could not join: %v", err)
	}

	placeFleet(room, game.SideA)
	placeFleet(room, game.SideB)

	r := lipgloss.DefaultRenderer()
	en := i18n.For(i18n.EN)
	ga := newGameModel(room, a, en, r)
	gb := newGameModel(room, b, en, r)

	if ga.phase != gameBattle || gb.phase != gameBattle {
		t.Fatalf("both players should be in battle, got %v and %v", ga.phase, gb.phase)
	}
	if !ga.snap.YourTurn || gb.snap.YourTurn {
		t.Fatal("SideA (Ali) should have the first turn")
	}

	// Ali fires at A1, where SideB's Carrier sits.
	ga.aim = game.Coord{Row: 0, Col: 0}
	updated, _ := ga.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	ga = updated.(gameModel)
	if ga.snap.YourTurn {
		t.Fatal("after firing, it should no longer be Ali's turn")
	}

	// Veli's update channel would wake it; simulate that with roomUpdateMsg.
	updated, _ = gb.Update(roomUpdateMsg{})
	gb = updated.(gameModel)
	if !gb.snap.YourTurn {
		t.Fatal("after Ali's shot it should be Veli's turn")
	}
	if got := gb.snap.You[0][0]; got != game.CellHit {
		t.Fatalf("Veli's board should show a hit at A1, got %v", got)
	}
}

// TestOpponentLeaveReturnsYouToWaiting checks that when the opponent leaves, the
// remaining player drops back to a fresh waiting room instead of being stuck.
func TestOpponentLeaveReturnsYouToWaiting(t *testing.T) {
	l := lobby.New(nil)
	a := lobby.NewHumanSeat("Ali")
	room := l.CreateRoom(a, "", false)
	b := lobby.NewHumanSeat("Veli")
	l.JoinByCode(room.Code, b, "")
	placeFleet(room, game.SideA)
	placeFleet(room, game.SideB)

	r := lipgloss.DefaultRenderer()
	gb := newGameModel(room, b, i18n.For(i18n.EN), r)

	// Ali leaves; Veli refreshes and should return to waiting, score wiped.
	l.Leave(room, a)
	updated, _ := gb.Update(roomUpdateMsg{})
	gb = updated.(gameModel)
	if gb.phase != gameWaiting {
		t.Fatalf("Veli should return to waiting, got phase=%v", gb.phase)
	}
	if gb.snap.YourScore != 0 {
		t.Fatalf("score should reset after opponent leaves, got %d", gb.snap.YourScore)
	}
}
