package math

type Vec2 struct {
	X int
	Y int
}

func (v Vec2) Eq(other Vec2) bool {
	return v.X == other.X && v.Y == other.Y
}

func (v Vec2) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

func Vec2Zero() Vec2 {
	return Vec2{X: 0, Y: 0}
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}
