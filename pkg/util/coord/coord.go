package coord

type Coord struct {
	X int
	Y int
}

func New(x, y int) Coord {
	return Coord{
		X: x,
		Y: y,
	}
}

func Plus(a, b Coord) Coord {
	return Coord{
		X: a.X + b.X,
		Y: a.Y + b.Y,
	}
}

func (c *Coord) Plus(other Coord) {
	*c = Plus(*c, other)
}
