package lobby

import (
	"testing"

	"github.com/ensardev/ssh-torpido/internal/game"
)

func TestScoreCountedOnceOnMatchEnd(t *testing.T) {
	r, _ := botRoomWithHuman()
	placeStandardFleet(r, game.SideA) // both placed -> battle
	r.match.Resign(game.SideB)        // the human (SideA) wins

	r.mu.Lock()
	r.scoreIfEndedLocked()
	r.scoreIfEndedLocked() // calling twice must not double count
	r.mu.Unlock()

	if r.score[game.SideA] != 1 {
		t.Fatalf("winner should have exactly 1 win, got %d", r.score[game.SideA])
	}
}

func TestBotRematchStartsFreshKeepingScore(t *testing.T) {
	r, human := botRoomWithHuman()
	placeStandardFleet(r, game.SideA)
	r.match.Resign(game.SideB)
	r.mu.Lock()
	r.scoreIfEndedLocked()
	r.mu.Unlock()

	side, _ := r.SideOf(human)
	r.RequestRematch(side) // bot opponent agrees immediately

	snap := r.Snapshot(side)
	if snap.MatchNo != 2 {
		t.Fatalf("rematch should bump matchNo to 2, got %d", snap.MatchNo)
	}
	if snap.Over {
		t.Fatal("the rematch should be a fresh, unfinished match")
	}
	if snap.YourScore != 1 {
		t.Fatalf("score should carry into the rematch, got %d", snap.YourScore)
	}
	if !snap.OppPlaced {
		t.Fatal("the bot should have re-placed its fleet for the rematch")
	}
}

func TestHumanRematchNeedsBothSides(t *testing.T) {
	r, a, b := humanRoomWithTwo()
	placeStandardFleet(r, game.SideA)
	placeStandardFleet(r, game.SideB)
	r.match.Resign(game.SideB)
	r.mu.Lock()
	r.scoreIfEndedLocked()
	r.mu.Unlock()

	sa, _ := r.SideOf(a)
	sb, _ := r.SideOf(b)

	r.RequestRematch(sa)
	if r.Snapshot(sa).MatchNo != 1 {
		t.Fatal("one side asking should not restart the match")
	}
	r.RequestRematch(sb)
	if r.Snapshot(sa).MatchNo != 2 {
		t.Fatal("both sides asking should restart the match")
	}
}

// A bot game that ends is scored for the win; leaving a human alone resets the
// series (covered in alone_test.go), so there is no forfeit-win to assert here.
