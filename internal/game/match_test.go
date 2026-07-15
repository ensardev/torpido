package game

import "testing"

// placeTinyFleet gives a side a single 1-square "ship" so matches end quickly.
func placeTinyFleet(m *Match, s Side, c Coord) {
	m.Place(s, "Raft", 1, []Coord{c})
	m.FinishPlacing(s)
}

func TestMatchStartsInPlacing(t *testing.T) {
	m := NewMatch()
	if m.Phase() != MatchPlacing {
		t.Fatalf("new match should be placing, got %v", m.Phase())
	}
	if _, ok := m.Winner(); ok {
		t.Fatal("new match should have no winner")
	}
}

func TestBattleStartsOnlyWhenBothPlaced(t *testing.T) {
	m := NewMatch()
	m.Place(SideA, "Raft", 1, []Coord{{0, 0}})
	m.FinishPlacing(SideA)
	if m.Phase() != MatchPlacing {
		t.Fatal("battle should not start until both sides are placed")
	}
	m.Place(SideB, "Raft", 1, []Coord{{9, 9}})
	m.FinishPlacing(SideB)
	if m.Phase() != MatchBattle {
		t.Fatalf("battle should start once both placed, got %v", m.Phase())
	}
	if m.Turn() != SideA {
		t.Fatalf("SideA should fire first, got %v", m.Turn())
	}
}

func TestFireRejectedOutOfTurn(t *testing.T) {
	m := NewMatch()
	placeTinyFleet(m, SideA, Coord{0, 0})
	placeTinyFleet(m, SideB, Coord{9, 9})
	// It is SideA's turn, so SideB firing must be rejected.
	if res, _ := m.Fire(SideB, Coord{0, 0}); res != FireInvalid {
		t.Fatalf("firing out of turn should be invalid, got %v", res)
	}
}

func TestTurnAlternatesOnMiss(t *testing.T) {
	m := NewMatch()
	placeTinyFleet(m, SideA, Coord{0, 0})
	placeTinyFleet(m, SideB, Coord{9, 9})
	if res, _ := m.Fire(SideA, Coord{5, 5}); res != FireMiss {
		t.Fatalf("expected miss, got %v", res)
	}
	if m.Turn() != SideB {
		t.Fatal("turn should pass to SideB after SideA's miss")
	}
}

func TestSinkingLastShipWinsTheMatch(t *testing.T) {
	m := NewMatch()
	placeTinyFleet(m, SideA, Coord{0, 0})
	placeTinyFleet(m, SideB, Coord{9, 9}) // SideB's only ship sits at J10
	res, _ := m.Fire(SideA, Coord{9, 9})
	if res != FireSunk {
		t.Fatalf("expected to sink SideB's raft, got %v", res)
	}
	if m.Phase() != MatchOver {
		t.Fatal("match should be over")
	}
	w, ok := m.Winner()
	if !ok || w != SideA {
		t.Fatalf("SideA should have won, got winner=%v ok=%v", w, ok)
	}
}

func TestResignHandsWinToOpponent(t *testing.T) {
	m := NewMatch()
	placeTinyFleet(m, SideA, Coord{0, 0})
	placeTinyFleet(m, SideB, Coord{9, 9})
	m.Resign(SideA)
	if m.Phase() != MatchOver {
		t.Fatal("match should be over after resign")
	}
	if w, ok := m.Winner(); !ok || w != SideB {
		t.Fatalf("SideB should win when SideA resigns, got winner=%v ok=%v", w, ok)
	}
}

func TestNoFiringAfterMatchOver(t *testing.T) {
	m := NewMatch()
	placeTinyFleet(m, SideA, Coord{0, 0})
	placeTinyFleet(m, SideB, Coord{9, 9})
	m.Fire(SideA, Coord{9, 9}) // SideA wins
	if res, _ := m.Fire(SideB, Coord{0, 0}); res != FireInvalid {
		t.Fatalf("no shots allowed after the match is over, got %v", res)
	}
}
