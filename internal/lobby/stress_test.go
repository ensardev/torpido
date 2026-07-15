package lobby

import (
	"sync"
	"testing"
)

// TestConcurrentChaosKeepsInvariant hammers the lobby from many goroutines doing
// joins, leaves, quick matches and list reads at once. Without a C compiler we
// can't use -race, but this still surfaces deadlocks (the test would hang and
// time out) and panics, and checks the bot invariant survives the storm.
func TestConcurrentChaosKeepsInvariant(t *testing.T) {
	l := New(nil)

	const workers = 40
	const rounds = 50
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				switch (id + i) % 3 {
				case 0:
					// Grab any idle bot room, then leave it.
					if code := botCodeForTier(l, botTiers[i%len(botTiers)]); code != "" {
						seat := NewHumanSeat("chaos")
						if r, err := l.JoinByCode(code, seat, ""); err == nil {
							l.Leave(r, seat)
						}
					}
				case 1:
					// Quick-match in and immediately bail out.
					seat := NewHumanSeat("chaos")
					r := l.QuickMatch(seat)
					l.Leave(r, seat)
				case 2:
					// Just read the public list.
					_ = l.PublicRooms()
				}
			}
		}(w)
	}
	wg.Wait()

	// After all the noise settles, the invariant must hold again.
	counts := idleBotTiers(l)
	for _, tier := range botTiers {
		if counts[tier] != 1 {
			t.Fatalf("after chaos expected 1 idle %s room, got %d", tier.Name(), counts[tier])
		}
	}

	// And no orphan human rooms should be left lying around.
	l.mu.Lock()
	humanRooms := 0
	for _, r := range l.rooms {
		if r.Kind == HumanRoom {
			humanRooms++
		}
	}
	l.mu.Unlock()
	if humanRooms != 0 {
		t.Fatalf("expected no leftover human rooms, got %d", humanRooms)
	}
}
