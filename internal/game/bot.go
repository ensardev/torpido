package game

import "math/rand"

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

// Bot decides where to fire. It uses the classic "hunt and target" strategy:
// fire randomly until it hits something, then work outwards from that hit
// before going back to random fire.
type Bot struct {
	rng     *rand.Rand
	targets []Coord       // squares queued to try after a hit (a small stack)
	tried   map[Coord]bool
}

// NewBot returns a bot with its own random source seeded by seed.
func NewBot(seed int64) *Bot {
	return &Bot{
		rng:   rand.New(rand.NewSource(seed)),
		tried: make(map[Coord]bool),
	}
}

// NextShot returns the next square the bot wants to fire at. It never returns a
// square it has already fired at.
func (bot *Bot) NextShot() Coord {
	// Target mode: follow up on a previous hit.
	for len(bot.targets) > 0 {
		c := bot.targets[len(bot.targets)-1]
		bot.targets = bot.targets[:len(bot.targets)-1]
		if c.Valid() && !bot.tried[c] {
			bot.tried[c] = true
			return c
		}
	}
	// Hunt mode: pick a random square we haven't tried yet.
	for {
		c := Coord{bot.rng.Intn(BoardSize), bot.rng.Intn(BoardSize)}
		if !bot.tried[c] {
			bot.tried[c] = true
			return c
		}
	}
}

// Report tells the bot how its last shot went so it can plan the next one.
// On a hit it queues the neighbouring squares to hunt down the rest of the ship.
func (bot *Bot) Report(c Coord, result FireResult) {
	if result != FireHit {
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
