package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/lobby"
)

func idleAdmiralCount(l *lobby.Lobby) int {
	n := 0
	for _, info := range l.PublicRooms() {
		if info.Kind == lobby.BotRoom && info.Tier == game.Admiral {
			n++
		}
	}
	return n
}

func admiralCode(l *lobby.Lobby) string {
	for _, info := range l.PublicRooms() {
		if info.Kind == lobby.BotRoom && info.Tier == game.Admiral {
			return info.Code
		}
	}
	return ""
}

// TestEnterAndLeaveBotRoom drives the root model through the whole navigation
// loop: lobby -> bot game -> back to lobby, and checks the room is freed and the
// bot invariant restored.
func TestEnterAndLeaveBotRoom(t *testing.T) {
	l := lobby.New()
	root := NewRoot(l, "sen", "", nil, lipgloss.DefaultRenderer())

	seat := lobby.NewHumanSeat("sen")
	room, err := l.JoinByCode(admiralCode(l), seat, "")
	if err != nil {
		t.Fatalf("joining the Admiral room failed: %v", err)
	}

	updated, _ := root.Update(enterRoomMsg{room: room, seat: seat})
	root = updated.(Root)
	if root.screen != rootGame {
		t.Fatal("root should be in game after entering a bot room")
	}

	updated, _ = root.Update(leaveGameMsg{})
	root = updated.(Root)
	if root.screen != rootLobby {
		t.Fatal("root should be back in the lobby after leaving")
	}
	if root.room != nil {
		t.Fatal("the room should be freed on leave")
	}
	if got := idleAdmiralCount(l); got != 1 {
		t.Fatalf("expected 1 idle Admiral room after leaving, got %d", got)
	}
}
