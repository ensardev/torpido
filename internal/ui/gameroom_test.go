package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
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
	l := lobby.New()
	a := lobby.NewHumanSeat("Ali")
	room := l.CreateRoom(a, "", false)
	b := lobby.NewHumanSeat("Veli")
	if _, err := l.JoinByCode(room.Code, b, ""); err != nil {
		t.Fatalf("Veli could not join: %v", err)
	}

	placeFleet(room, game.SideA)
	placeFleet(room, game.SideB)

	r := lipgloss.DefaultRenderer()
	ga := newGameModel(room, a, r)
	gb := newGameModel(room, b, r)

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

// TestOpponentForfeitOnLeave checks that when one player leaves, the other is
// shown as the winner.
func TestOpponentForfeitOnLeave(t *testing.T) {
	l := lobby.New()
	a := lobby.NewHumanSeat("Ali")
	room := l.CreateRoom(a, "", false)
	b := lobby.NewHumanSeat("Veli")
	l.JoinByCode(room.Code, b, "")
	placeFleet(room, game.SideA)
	placeFleet(room, game.SideB)

	r := lipgloss.DefaultRenderer()
	gb := newGameModel(room, b, r)

	// Ali leaves; Veli refreshes and should see a win.
	l.Leave(room, a)
	updated, _ := gb.Update(roomUpdateMsg{})
	gb = updated.(gameModel)
	if gb.phase != gameOver || !gb.snap.YouWon {
		t.Fatalf("Veli should win by forfeit, got phase=%v won=%v", gb.phase, gb.snap.YouWon)
	}
}
