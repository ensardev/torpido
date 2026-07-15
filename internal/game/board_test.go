package game

import (
	"math/rand"
	"testing"
)

func TestFireMissHitSunk(t *testing.T) {
	b := NewBoard()
	// A destroyer of size 2 along the top-left.
	b.Place("Destroyer", 2, []Coord{{0, 0}, {0, 1}})

	if res, _ := b.Fire(Coord{5, 5}); res != FireMiss {
		t.Fatalf("expected miss on empty water, got %v", res)
	}
	if res, _ := b.Fire(Coord{0, 0}); res != FireHit {
		t.Fatalf("expected hit on first ship square, got %v", res)
	}
	res, ship := b.Fire(Coord{0, 1})
	if res != FireSunk {
		t.Fatalf("expected sunk on last ship square, got %v", res)
	}
	if ship == nil || ship.Name != "Destroyer" {
		t.Fatalf("expected sunk ship to be the Destroyer, got %v", ship)
	}
}

func TestFireInvalidWhenRepeatedOrOffBoard(t *testing.T) {
	b := NewBoard()
	b.Fire(Coord{2, 2})
	if res, _ := b.Fire(Coord{2, 2}); res != FireInvalid {
		t.Fatalf("expected invalid when firing twice at the same square, got %v", res)
	}
	if res, _ := b.Fire(Coord{-1, 0}); res != FireInvalid {
		t.Fatalf("expected invalid when firing off the board, got %v", res)
	}
}

func TestAllSunk(t *testing.T) {
	b := NewBoard()
	b.Place("Destroyer", 2, []Coord{{0, 0}, {0, 1}})
	if b.AllSunk() {
		t.Fatal("fleet should not be sunk before any shots")
	}
	b.Fire(Coord{0, 0})
	b.Fire(Coord{0, 1})
	if !b.AllSunk() {
		t.Fatal("fleet should be sunk after every square is hit")
	}
}

func TestCanPlaceRejectsOverlap(t *testing.T) {
	b := NewBoard()
	b.Place("Cruiser", 3, ShipCoords(Coord{0, 0}, 3, Horizontal))
	if b.CanPlace(ShipCoords(Coord{0, 2}, 3, Vertical)) {
		t.Fatal("overlapping placement should be rejected")
	}
	if !b.CanPlace(ShipCoords(Coord{5, 5}, 3, Vertical)) {
		t.Fatal("clear placement should be allowed")
	}
}

func TestBotNeverRepeatsAShot(t *testing.T) {
	bot := NewBot(Admiral, 1)
	seen := make(map[Coord]bool)
	for i := 0; i < BoardSize*BoardSize; i++ {
		c := bot.NextShot()
		if seen[c] {
			t.Fatalf("bot fired at %v twice", c)
		}
		seen[c] = true
	}
}

func TestRandomPlacementFitsWholeFleet(t *testing.T) {
	b := NewBoard()
	RandomPlacement(b, StandardFleet, rand.New(rand.NewSource(42)))
	cells := 0
	for _, s := range b.Ships {
		cells += len(s.Coords)
	}
	if cells != 17 {
		t.Fatalf("expected 17 occupied squares for the standard fleet, got %d", cells)
	}
}
