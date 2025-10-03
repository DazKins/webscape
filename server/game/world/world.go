package world

import (
	"math/rand"
	"webscape/server/util"
)

type WallDirection int

const (
	WallDirectionUp WallDirection = iota
	WallDirectionDown
	WallDirectionLeft
	WallDirectionRight
)

type World struct {
	sizeX int
	sizeY int

	walls [][]*WallDirection
}

func NewWorld(sizeX int, sizeY int) *World {
	walls := make([][]*WallDirection, sizeX)
	for i := range walls {
		walls[i] = make([]*WallDirection, sizeY)
		for j := range walls[i] {
			if rand.Intn(100) < 10 {
				walls[i][j] = util.Ptr(WallDirectionUp)
			}
		}
	}

	return &World{
		sizeX: sizeX,
		sizeY: sizeY,
		walls: walls,
	}
}

func (w *World) GetSizeX() int {
	return w.sizeX
}

func (w *World) GetSizeY() int {
	return w.sizeY
}

func (w *World) GetWalls() [][]*WallDirection {
	return w.walls
}
