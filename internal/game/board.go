// Package game holds torpido's pure game rules: the board, ships, shots and
// turn resolution. It knows nothing about terminals, SSH or rendering, so the
// same logic drives the local UI now and the SSH server later.
package game

// BoardSize is the width and height of the playing grid (10x10, A-J / 1-10).
const BoardSize = 10

// Cell is what a single square shows when rendered.
type Cell int

const (
	CellEmpty Cell = iota // open water, nothing known
	CellShip              // a ship occupies this square (only shown to its owner)
	CellHit               // a shot landed on a ship here
	CellMiss              // a shot landed on empty water here
	CellSunk              // this square belongs to a fully sunk ship
)

// Orientation is how a ship lies on the board.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Coord is a single square, zero-indexed (Row 0-9, Col 0-9).
type Coord struct {
	Row, Col int
}

// Valid reports whether the coord sits inside the board.
func (c Coord) Valid() bool {
	return c.Row >= 0 && c.Row < BoardSize && c.Col >= 0 && c.Col < BoardSize
}

// Ship is one vessel placed on a board.
type Ship struct {
	Name   string
	Size   int
	Coords []Coord
	Hits   int
}

// Sunk reports whether every square of the ship has been hit.
func (s *Ship) Sunk() bool {
	return s.Hits >= s.Size
}

// Board is one player's grid: the ships on it and the shots fired at it.
type Board struct {
	Ships    []*Ship
	Shots    map[Coord]bool  // squares that have been fired at
	occupied map[Coord]*Ship // fast lookup of which ship (if any) sits on a square
}

// NewBoard returns an empty board ready for ships to be placed.
func NewBoard() *Board {
	return &Board{
		Shots:    make(map[Coord]bool),
		occupied: make(map[Coord]*Ship),
	}
}

// ShipCoords returns the squares a ship of the given size would cover starting
// at start and running in the given orientation.
func ShipCoords(start Coord, size int, o Orientation) []Coord {
	coords := make([]Coord, size)
	for i := 0; i < size; i++ {
		if o == Horizontal {
			coords[i] = Coord{start.Row, start.Col + i}
		} else {
			coords[i] = Coord{start.Row + i, start.Col}
		}
	}
	return coords
}

// CanPlace reports whether every square is on the board and unoccupied.
func (b *Board) CanPlace(coords []Coord) bool {
	for _, c := range coords {
		if !c.Valid() {
			return false
		}
		if _, taken := b.occupied[c]; taken {
			return false
		}
	}
	return true
}

// Place puts a ship on the board. It returns false (and changes nothing) if the
// squares are off the board or overlap another ship.
func (b *Board) Place(name string, size int, coords []Coord) bool {
	if !b.CanPlace(coords) {
		return false
	}
	s := &Ship{Name: name, Size: size, Coords: append([]Coord(nil), coords...)}
	b.Ships = append(b.Ships, s)
	for _, c := range coords {
		b.occupied[c] = s
	}
	return true
}

// FireResult is the outcome of firing at a square.
type FireResult int

const (
	FireInvalid FireResult = iota // off the board or already fired at
	FireMiss                      // hit open water
	FireHit                       // hit a ship (still afloat)
	FireSunk                      // hit a ship and sank it
)

// Fire resolves a shot at c. On a hit it returns the ship that was struck.
func (b *Board) Fire(c Coord) (FireResult, *Ship) {
	if !c.Valid() || b.Shots[c] {
		return FireInvalid, nil
	}
	b.Shots[c] = true
	ship, ok := b.occupied[c]
	if !ok {
		return FireMiss, nil
	}
	ship.Hits++
	if ship.Sunk() {
		return FireSunk, ship
	}
	return FireHit, ship
}

// AllSunk reports whether every ship on the board has been sunk.
func (b *Board) AllSunk() bool {
	if len(b.Ships) == 0 {
		return false
	}
	for _, s := range b.Ships {
		if !s.Sunk() {
			return false
		}
	}
	return true
}

// StateAt returns what a square should display. revealShips controls whether
// un-hit ships are visible (true for your own board, false for the enemy's).
func (b *Board) StateAt(c Coord, revealShips bool) Cell {
	fired := b.Shots[c]
	ship, hasShip := b.occupied[c]
	switch {
	case fired && hasShip && ship.Sunk():
		return CellSunk
	case fired && hasShip:
		return CellHit
	case fired:
		return CellMiss
	case hasShip && revealShips:
		return CellShip
	default:
		return CellEmpty
	}
}
