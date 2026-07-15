package lobby

import (
	"math/rand"
	"testing"

	"github.com/ensardev/ssh-torpido/internal/game"
)

// placeStandardFleet lays a full fleet in a known corner so a side finishes
// placing deterministically in tests.
func placeStandardFleet(r *Room, side game.Side) {
	rows := [][2]int{{0, 5}, {1, 4}, {2, 3}, {3, 3}, {4, 2}} // {row, size} per ship
	names := []string{"Carrier", "Battleship", "Cruiser", "Submarine", "Destroyer"}
	for i, rs := range rows {
		row, size := rs[0], rs[1]
		coords := game.ShipCoords(game.Coord{Row: row, Col: 0}, size, game.Horizontal)
		r.PlaceShip(side, names[i], size, coords)
	}
	r.FinishPlacing(side)
}

func botRoomWithHuman() (*Room, *Seat) {
	r := &Room{Kind: BotRoom, Tier: game.Admiral, match: game.NewMatch()}
	r.seats[1] = &Seat{Name: "Amiral", bot: game.NewBot(game.Admiral, 1)}
	game.RandomPlacement(r.match.Board(game.SideB), game.StandardFleet, rand.New(rand.NewSource(1)))
	r.match.FinishPlacing(game.SideB)
	human := NewHumanSeat("Ali")
	r.seats[0] = human
	return r, human
}

func humanRoomWithTwo() (*Room, *Seat, *Seat) {
	r := &Room{Kind: HumanRoom, match: game.NewMatch()}
	a, b := NewHumanSeat("Ali"), NewHumanSeat("Veli")
	r.seats[0], r.seats[1] = a, b
	return r, a, b
}

func drain(seat *Seat) bool {
	select {
	case <-seat.updates:
		return true
	default:
		return false
	}
}

func TestSideOf(t *testing.T) {
	r, a, b := humanRoomWithTwo()
	if s, ok := r.SideOf(a); !ok || s != game.SideA {
		t.Fatalf("seat A should be SideA, got %v ok=%v", s, ok)
	}
	if s, ok := r.SideOf(b); !ok || s != game.SideB {
		t.Fatalf("seat B should be SideB, got %v ok=%v", s, ok)
	}
	if _, ok := r.SideOf(NewHumanSeat("stranger")); ok {
		t.Fatal("a seat not in the room should report not found")
	}
}

func TestBattleBeginsWhenBothPlaced(t *testing.T) {
	r, human := botRoomWithHuman()
	snap := r.Snapshot(game.SideA)
	if snap.Phase != game.MatchPlacing {
		t.Fatal("should still be placing before the human places")
	}
	if !snap.OppPlaced {
		t.Fatal("the bot should already be placed")
	}
	placeStandardFleet(r, game.SideA)
	_ = human
	snap = r.Snapshot(game.SideA)
	if snap.Phase != game.MatchBattle {
		t.Fatalf("battle should start once the human places, got %v", snap.Phase)
	}
	if !snap.YourTurn {
		t.Fatal("the human (SideA) should fire first")
	}
}

func TestFireWakesOpponent(t *testing.T) {
	r, a, b := humanRoomWithTwo()
	placeStandardFleet(r, game.SideA)
	placeStandardFleet(r, game.SideB)
	drain(a)
	drain(b)

	// SideA fires; SideB must be woken.
	r.Fire(game.SideA, game.Coord{Row: 9, Col: 9})
	if !drain(b) {
		t.Fatal("opponent should be notified after a shot")
	}
	// Out-of-turn shot by SideA now (it's B's turn) changes nothing / no wake.
	drain(a)
	drain(b)
	if res, _ := r.Fire(game.SideA, game.Coord{Row: 8, Col: 8}); res != game.FireInvalid {
		t.Fatalf("SideA firing out of turn should be invalid, got %v", res)
	}
	if drain(b) {
		t.Fatal("an invalid shot should not wake the opponent")
	}
}

func TestPlayBotTurnOnlyWhenBotsTurn(t *testing.T) {
	r, _ := botRoomWithHuman()
	placeStandardFleet(r, game.SideA)

	// It's the human's turn first, so the bot must not act.
	if r.PlayBotTurn() {
		t.Fatal("bot should not play on the human's turn")
	}
	// Human fires a miss, passing the turn to the bot.
	r.Fire(game.SideA, game.Coord{Row: 9, Col: 9})
	if !r.PlayBotTurn() {
		t.Fatal("bot should play once it is its turn")
	}
	// After the bot plays, the turn returns to the human (unless someone won).
	snap := r.Snapshot(game.SideA)
	if !snap.Over && !snap.YourTurn {
		t.Fatal("turn should return to the human after the bot's shot")
	}
}

func TestSnapshotHidesEnemyShips(t *testing.T) {
	r, _, _ := humanRoomWithTwo()
	placeStandardFleet(r, game.SideA)
	placeStandardFleet(r, game.SideB)

	snap := r.Snapshot(game.SideA)
	// Your own board should reveal ships; the enemy board should not.
	yourShips, enemyShips := 0, 0
	for row := 0; row < game.BoardSize; row++ {
		for col := 0; col < game.BoardSize; col++ {
			if snap.You[row][col] == game.CellShip {
				yourShips++
			}
			if snap.Enemy[row][col] == game.CellShip {
				enemyShips++
			}
		}
	}
	if yourShips != 17 {
		t.Fatalf("your board should show all 17 ship squares, got %d", yourShips)
	}
	if enemyShips != 0 {
		t.Fatalf("enemy ships must stay hidden, got %d visible", enemyShips)
	}
}
