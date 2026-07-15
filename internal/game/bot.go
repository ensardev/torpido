package game

import "math/rand"

// Difficulty is how smart a bot plays.
type Difficulty int

const (
	Rookie  Difficulty = iota // fires at random, never follows up on a hit
	Admiral                    // hunt & target: works outwards from each hit
	SeaWolf                    // hunt & target, but hunts on a parity grid to find ships faster
)

// Name is the bot's display name for its difficulty.
func (d Difficulty) Name() string {
	switch d {
	case Rookie:
		return "Acemi Er"
	case Admiral:
		return "Amiral"
	case SeaWolf:
		return "Deniz Kurdu"
	default:
		return "Bot"
	}
}

// RandomPlacement drops every ship of the fleet onto the board at random,
// retrying until each one fits without overlapping.
func RandomPlacement(b *Board, fleet []ShipType, rng *rand.Rand) {
	for _, st := range fleet {
		for {
			o := Orientation(rng.Intn(2))
			var start Coord
			if o == Horizontal {
				start = Coord{rng.Intn(BoardSize), rng.Intn(BoardSize - st.Size + 1)}
			} else {
				start = Coord{rng.Intn(BoardSize - st.Size + 1), rng.Intn(BoardSize)}
			}
			if b.Place(st.Name, st.Size, ShipCoords(start, st.Size, o)) {
				break
			}
		}
	}
}

// Bot decides where to fire. Rookie fires blindly; Admiral and SeaWolf use the
// classic "hunt and target" idea — once they hit something they work outwards to
// finish the ship before going back to hunting.
type Bot struct {
	difficulty Difficulty
	rng        *rand.Rand
	targets    []Coord // squares queued to try after a hit
	tried      map[Coord]bool
}

// NewBot returns a bot of the given difficulty with its own random source.
func NewBot(d Difficulty, seed int64) *Bot {
	return &Bot{
		difficulty: d,
		rng:        rand.New(rand.NewSource(seed)),
		tried:      make(map[Coord]bool),
	}
}

// Difficulty reports how the bot plays.
func (bot *Bot) Difficulty() Difficulty { return bot.difficulty }

// NextShot returns the next square the bot wants to fire at. It never returns a
// square it has already fired at.
func (bot *Bot) NextShot() Coord {
	// Target mode (Admiral/SeaWolf): follow up on a previous hit.
	if bot.difficulty != Rookie {
		for len(bot.targets) > 0 {
			c := bot.targets[len(bot.targets)-1]
			bot.targets = bot.targets[:len(bot.targets)-1]
			if c.Valid() && !bot.tried[c] {
				bot.mark(c)
				return c
			}
		}
	}
	// Hunt mode: SeaWolf prefers a parity grid (every other square), which is
	// enough to catch any ship of length >= 2 with half the shots.
	if bot.difficulty == SeaWolf {
		if c, ok := bot.randomUntried(true); ok {
			bot.mark(c)
			return c
		}
	}
	c, _ := bot.randomUntried(false)
	bot.mark(c)
	return c
}

// Report tells the bot how its last shot went so it can plan the next one. On a
// hit, hunting bots queue the neighbouring squares to finish off the ship.
func (bot *Bot) Report(c Coord, result FireResult) {
	if bot.difficulty == Rookie || result != FireHit {
		return
	}
	for _, n := range []Coord{
		{c.Row - 1, c.Col},
		{c.Row + 1, c.Col},
		{c.Row, c.Col - 1},
		{c.Row, c.Col + 1},
	} {
		if n.Valid() && !bot.tried[n] {
			bot.targets = append(bot.targets, n)
		}
	}
}

func (bot *Bot) mark(c Coord) { bot.tried[c] = true }

// randomUntried returns a random square the bot hasn't fired at. When parityOnly
// is set it only considers squares where (row+col) is even, and reports false if
// none of those remain.
func (bot *Bot) randomUntried(parityOnly bool) (Coord, bool) {
	var candidates []Coord
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			coord := Coord{Row: r, Col: c}
			if bot.tried[coord] {
				continue
			}
			if parityOnly && (r+c)%2 != 0 {
				continue
			}
			candidates = append(candidates, coord)
		}
	}
	if len(candidates) == 0 {
		return Coord{}, false
	}
	return candidates[bot.rng.Intn(len(candidates))], true
}
