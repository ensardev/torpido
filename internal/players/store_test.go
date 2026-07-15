package players

import (
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "stats.json"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	return s
}

func TestRecordResultAndPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stats.json")
	s, _ := Open(path)
	s.Ensure("fp1", "Ali")
	s.RecordResult("fp1", true)
	s.RecordResult("fp1", true)
	s.RecordResult("fp1", false)

	// Reopen from disk: the tally must have survived.
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	r, ok := s2.Get("fp1")
	if !ok {
		t.Fatal("record should persist to disk")
	}
	if r.Wins != 2 || r.Losses != 1 {
		t.Fatalf("expected 2-1, got %d-%d", r.Wins, r.Losses)
	}
	if r.Nick != "Ali" {
		t.Fatalf("nick should persist, got %q", r.Nick)
	}
}

func TestTopAndRank(t *testing.T) {
	s := tempStore(t)
	s.Ensure("a", "Ali")
	s.Ensure("b", "Veli")
	s.Ensure("c", "Can")
	for i := 0; i < 5; i++ {
		s.RecordResult("b", true)
	}
	for i := 0; i < 3; i++ {
		s.RecordResult("a", true)
	}
	// c has 0 wins.

	top := s.Top(2)
	if len(top) != 2 || top[0].Nick != "Veli" || top[1].Nick != "Ali" {
		t.Fatalf("leaderboard order wrong: %+v", top)
	}
	if r := s.Rank("b"); r != 1 {
		t.Fatalf("Veli should rank 1, got %d", r)
	}
	if r := s.Rank("a"); r != 2 {
		t.Fatalf("Ali should rank 2, got %d", r)
	}
	if r := s.Rank("c"); r != 3 {
		t.Fatalf("Can should rank 3, got %d", r)
	}
}

func TestNickUniqueness(t *testing.T) {
	s := tempStore(t)
	s.Ensure("a", "Ali")
	s.Ensure("b", "Veli")

	if !s.SetNick("a", "Kaptan") {
		t.Fatal("a should be able to claim a free nick")
	}
	if s.SetNick("b", "kaptan") {
		t.Fatal("b should not be able to claim a taken nick (case-insensitive)")
	}
	if !s.SetNick("a", "Kaptan") {
		t.Fatal("a re-claiming its own nick should succeed")
	}
	if s.NickTaken("a", "Kaptan") {
		t.Fatal("your own nick should not count as taken by someone else")
	}
	if !s.NickTaken("b", "Kaptan") {
		t.Fatal("someone else's nick should read as taken")
	}
}
