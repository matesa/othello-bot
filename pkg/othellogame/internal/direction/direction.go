package direction

const COUNT = 8

type Direction int

const (
	NORTH_WEST Direction = iota
	NORTH
	NORTH_EAST
	WEST
	EAST
	SOUTH_WEST
	SOUTH
	SOUTH_EAST
)
