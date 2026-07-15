// Package players is torpido's persistent record of who has played: each SSH
// public key maps to a nickname and a win/loss tally that survives reconnects
// and server restarts. It is the "accountless account" — no passwords, identity
// is just your SSH key.
package players

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"sync"
)

// Record is one player's persistent stats.
type Record struct {
	Fingerprint string `json:"fp"`
	Nick        string `json:"nick"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
}

// Store keeps all records, backed by a JSON file on disk.
type Store struct {
	mu   sync.Mutex
	path string
	recs map[string]*Record
}

// Open loads the store from path, creating an empty one if the file is missing.
func Open(path string) (*Store, error) {
	s := &Store{path: path, recs: map[string]*Record{}}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		var list []*Record
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		for _, r := range list {
			s.recs[r.Fingerprint] = r
		}
	case !os.IsNotExist(err):
		return nil, err
	}
	return s, nil
}

func (s *Store) saveLocked() {
	list := make([]*Record, 0, len(s.recs))
	for _, r := range s.recs {
		list = append(list, r)
	}
	if data, err := json.MarshalIndent(list, "", "  "); err == nil {
		_ = os.WriteFile(s.path, data, 0o644) // best effort; stats are not critical
	}
}

// Ensure returns the record for fp, creating it with defaultNick when new.
func (s *Store) Ensure(fp, defaultNick string) Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.recs[fp]
	if !ok {
		r = &Record{Fingerprint: fp, Nick: defaultNick}
		s.recs[fp] = r
		s.saveLocked()
	}
	return *r
}

// Get returns the record for fp, if it exists.
func (s *Store) Get(fp string) (Record, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.recs[fp]; ok {
		return *r, true
	}
	return Record{}, false
}

// RecordResult adds a win or a loss to fp's tally.
func (s *Store) RecordResult(fp string, won bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.recs[fp]
	if !ok {
		return
	}
	if won {
		r.Wins++
	} else {
		r.Losses++
	}
	s.saveLocked()
}

// Top returns the n players with the most wins (ties broken by fewer losses).
func (s *Store) Top(n int) []Record {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]Record, 0, len(s.recs))
	for _, r := range s.recs {
		list = append(list, *r)
	}
	sortByWins(list)
	if len(list) > n {
		list = list[:n]
	}
	return list
}

// Rank returns fp's 1-based position on the leaderboard (0 if unknown).
func (s *Store) Rank(fp string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	me, ok := s.recs[fp]
	if !ok {
		return 0
	}
	rank := 1
	for _, r := range s.recs {
		if r.Fingerprint == fp {
			continue
		}
		if r.Wins > me.Wins || (r.Wins == me.Wins && r.Losses < me.Losses) {
			rank++
		}
	}
	return rank
}

// NickTaken reports whether some other player already uses nick.
func (s *Store) NickTaken(fp, nick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.recs {
		if r.Fingerprint != fp && strings.EqualFold(r.Nick, nick) {
			return true
		}
	}
	return false
}

// SetNick claims nick for fp, returning false if another player already has it.
func (s *Store) SetNick(fp, nick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.recs {
		if r.Fingerprint != fp && strings.EqualFold(r.Nick, nick) {
			return false
		}
	}
	r, ok := s.recs[fp]
	if !ok {
		r = &Record{Fingerprint: fp}
		s.recs[fp] = r
	}
	r.Nick = nick
	s.saveLocked()
	return true
}

func sortByWins(list []Record) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Wins != list[j].Wins {
			return list[i].Wins > list[j].Wins
		}
		if list[i].Losses != list[j].Losses {
			return list[i].Losses < list[j].Losses
		}
		return list[i].Nick < list[j].Nick
	})
}
