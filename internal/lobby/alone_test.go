package lobby

import (
	"testing"

	"github.com/ensardev/ssh-torpido/internal/game"
)

// TestOpponentLeaveReturnsHostToWaiting reproduces the "zombie room" case: after
// an opponent joins, plays and leaves, the remaining player's room must reset to
// a fresh waiting room (score wiped) and be re-listed for a new opponent.
func TestOpponentLeaveReturnsHostToWaiting(t *testing.T) {
	l := New()
	a := NewHumanSeat("Ali")
	room := l.CreateRoom(a, "", false) // public
	b := NewHumanSeat("Veli")
	if _, err := l.JoinByCode(room.Code, b, ""); err != nil {
		t.Fatal(err)
	}
	placeStandardFleet(room, game.SideA)
	placeStandardFleet(room, game.SideB)
	room.Fire(game.SideA, game.Coord{Row: 0, Col: 0}) // play a bit

	l.Leave(room, b) // Veli leaves; Ali is now alone

	sa, _ := room.SideOf(a)
	snap := room.Snapshot(sa)
	if snap.MatchNo != 2 {
		t.Fatalf("room should reset to a new match, matchNo=%d", snap.MatchNo)
	}
	if snap.YourScore != 0 || snap.OppScore != 0 {
		t.Fatalf("score should be wiped, got %d-%d", snap.YourScore, snap.OppScore)
	}
	if snap.OppPresent {
		t.Fatal("opponent should be gone")
	}

	listed := false
	for _, info := range l.PublicRooms() {
		if info.Code == room.Code {
			listed = true
		}
	}
	if !listed {
		t.Fatal("the room should be re-listed for a new opponent")
	}
}
