package util

type Coord struct {
	X int
	Y int
}

func NewCoord(x, y int) Coord {
	return Coord{
		X: x,
		Y: y,
	}
}

type CoordSet struct {
	m map[Coord]struct{}
}

func NewCoordSet() CoordSet {
	return CoordSet{
		m: make(map[Coord]struct{}),
	}
}

func (set *CoordSet) Clear() {
	for key := range set.m {
		delete(set.m, key)
	}
}

func (set *CoordSet) Insert(c Coord) {
	set.m[c] = struct{}{}
}

func (set *CoordSet) Contains(c Coord) bool {
	_, present := set.m[c]
	return present
}

func (set *CoordSet) IsEmpty() bool {
	return len(set.m) == 0
}
