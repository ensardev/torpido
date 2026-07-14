package game

// ShipType is a kind of ship in the fleet: a name and how many squares it fills.
type ShipType struct {
	Name string
	Size int
}

// StandardFleet is the classic five-ship line-up (17 squares in total).
var StandardFleet = []ShipType{
	{"Carrier", 5},
	{"Battleship", 4},
	{"Cruiser", 3},
	{"Submarine", 3},
	{"Destroyer", 2},
}
