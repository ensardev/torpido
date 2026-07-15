package lobby

import (
	"testing"

	"github.com/ensardev/torpido/internal/game"
)

// idleBotTiers returns how many idle bot rooms exist per tier in the public list.
func idleBotTiers(l *Lobby) map[game.Difficulty]int {
	counts := map[game.Difficulty]int{}
	for _, info := range l.PublicRooms() {
		if info.Kind == BotRoom {
			counts[info.Tier]++
		}
	}
	return counts
}

func botCodeForTier(l *Lobby, tier game.Difficulty) string {
	for _, info := range l.PublicRooms() {
		if info.Kind == BotRoom && info.Tier == tier {
			return info.Code
		}
	}
	return ""
}

func assertOneIdlePerTier(t *testing.T, l *Lobby) {
	t.Helper()
	counts := idleBotTiers(l)
	for _, tier := range botTiers {
		if counts[tier] != 1 {
			t.Fatalf("expected exactly 1 idle %s room, got %d", tier.Name(), counts[tier])
		}
	}
}

func TestNewLobbyHasOneIdleBotRoomPerTier(t *testing.T) {
	l := New()
	assertOneIdlePerTier(t, l)
}

func TestJoiningBotRoomTopsUpItsTier(t *testing.T) {
	l := New()
	code := botCodeForTier(l, game.Admiral)
	if _, err := l.JoinByCode(code, NewHumanSeat("Ali"), ""); err != nil {
		t.Fatalf("joining the Admiral bot room failed: %v", err)
	}
	// The joined room is now occupied and off the list, but a fresh idle Admiral
	// room must have replaced it — still one idle per tier.
	assertOneIdlePerTier(t, l)
	// And the occupied room's code is no longer offered as idle.
	if botCodeForTier(l, game.Admiral) == code {
		t.Fatal("the occupied Admiral room should no longer be listed as idle")
	}
}

func TestLeavingBotRoomRemovesItAndRestoresInvariant(t *testing.T) {
	l := New()
	code := botCodeForTier(l, game.SeaWolf)
	seat := NewHumanSeat("Ali")
	r, err := l.JoinByCode(code, seat, "")
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}
	l.Leave(r, seat)
	assertOneIdlePerTier(t, l)
}

func TestManyJoinsAndLeavesDoNotLeakBotRooms(t *testing.T) {
	l := New()
	for i := 0; i < 20; i++ {
		tier := botTiers[i%len(botTiers)]
		code := botCodeForTier(l, tier)
		seat := NewHumanSeat("player")
		r, err := l.JoinByCode(code, seat, "")
		if err != nil {
			t.Fatalf("join %d failed: %v", i, err)
		}
		l.Leave(r, seat)
	}
	assertOneIdlePerTier(t, l)
}

func TestCreateAndJoinHumanRoom(t *testing.T) {
	l := New()
	host := NewHumanSeat("Ali")
	r := l.CreateRoom(host, "", false)
	if r.Code == "" {
		t.Fatal("created room should have a code")
	}
	if _, err := l.JoinByCode(r.Code, NewHumanSeat("Veli"), ""); err != nil {
		t.Fatalf("second player should be able to join: %v", err)
	}
	if _, err := l.JoinByCode(r.Code, NewHumanSeat("Can"), ""); err != ErrRoomFull {
		t.Fatalf("third player should get ErrRoomFull, got %v", err)
	}
}

func TestPasswordProtectedRoom(t *testing.T) {
	l := New()
	r := l.CreateRoom(NewHumanSeat("Ali"), "1234", true)
	if _, err := l.JoinByCode(r.Code, NewHumanSeat("Veli"), "0000"); err != ErrBadPassword {
		t.Fatalf("wrong password should be rejected, got %v", err)
	}
	if _, err := l.JoinByCode(r.Code, NewHumanSeat("Veli"), "1234"); err != nil {
		t.Fatalf("correct password should be accepted, got %v", err)
	}
}

func TestPrivateRoomNotListed(t *testing.T) {
	l := New()
	r := l.CreateRoom(NewHumanSeat("Ali"), "", true)
	for _, info := range l.PublicRooms() {
		if info.Code == r.Code {
			t.Fatal("private room should not appear in the public list")
		}
	}
}

func TestQuickMatchPairsTwoPlayers(t *testing.T) {
	l := New()
	first := l.QuickMatch(NewHumanSeat("Ali"))
	second := l.QuickMatch(NewHumanSeat("Veli"))
	if first.Code != second.Code {
		t.Fatal("the second quick-match player should join the first player's room")
	}
}

func TestReconcilePrunesExtraIdleRooms(t *testing.T) {
	l := New()
	// Inject a duplicate idle Admiral room, then reconcile it away.
	l.mu.Lock()
	l.createBotRoomLocked(game.Admiral)
	l.mu.Unlock()
	if idleBotTiers(l)[game.Admiral] != 2 {
		t.Fatal("setup: expected 2 idle Admiral rooms before reconcile")
	}
	l.mu.Lock()
	l.reconcileLocked()
	l.mu.Unlock()
	assertOneIdlePerTier(t, l)
}
