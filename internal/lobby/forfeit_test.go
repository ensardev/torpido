package lobby

import (
	"path/filepath"
	"testing"

	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/players"
)

func storedLobby(t *testing.T) (*Lobby, *players.Store) {
	t.Helper()
	store, err := players.Open(filepath.Join(t.TempDir(), "stats.json"))
	if err != nil {
		t.Fatal(err)
	}
	return New(store), store
}

func humanMatch(t *testing.T, l *Lobby, store *players.Store) *Room {
	t.Helper()
	a := NewHumanSeat("Ali")
	a.Fingerprint = "fpA"
	room := l.CreateRoom(a, "", false)
	b := NewHumanSeat("Veli")
	b.Fingerprint = "fpB"
	if _, err := l.JoinByCode(room.Code, b, ""); err != nil {
		t.Fatal(err)
	}
	store.Ensure("fpA", "Ali")
	store.Ensure("fpB", "Veli")
	return room
}

func TestBattleForfeitRecordsWinAndLoss(t *testing.T) {
	l, store := storedLobby(t)
	room := humanMatch(t, l, store)
	placeStandardFleet(room, game.SideA)
	placeStandardFleet(room, game.SideB) // battle begins

	// Veli leaves mid-battle: forfeit.
	seatB := room.seats[game.SideB]
	l.Leave(room, seatB)

	ra, _ := store.Get("fpA")
	rb, _ := store.Get("fpB")
	if ra.Wins != 1 {
		t.Fatalf("the remaining player should gain a win, got %d", ra.Wins)
	}
	if rb.Losses != 1 {
		t.Fatalf("the forfeiter should take a loss, got %d", rb.Losses)
	}
}

func TestPlacementLeaveDoesNotCount(t *testing.T) {
	l, store := storedLobby(t)
	room := humanMatch(t, l, store)
	placeStandardFleet(room, game.SideA) // only Ali placed; still in placement

	seatB := room.seats[game.SideB]
	l.Leave(room, seatB) // Veli leaves during placement

	ra, _ := store.Get("fpA")
	rb, _ := store.Get("fpB")
	if ra.Wins != 0 || rb.Losses != 0 {
		t.Fatalf("no W/L during placement, got A wins=%d B losses=%d", ra.Wins, rb.Losses)
	}
}

func TestNormalMatchEndRecordsWinAndLoss(t *testing.T) {
	l, store := storedLobby(t)
	room := humanMatch(t, l, store)
	placeStandardFleet(room, game.SideA)
	placeStandardFleet(room, game.SideB)

	room.match.Resign(game.SideB) // as if Ali sank Veli's fleet
	room.mu.Lock()
	room.scoreIfEndedLocked()
	room.mu.Unlock()

	ra, _ := store.Get("fpA")
	rb, _ := store.Get("fpB")
	if ra.Wins != 1 || rb.Losses != 1 {
		t.Fatalf("winner/loser should be recorded, got A wins=%d B losses=%d", ra.Wins, rb.Losses)
	}
}
