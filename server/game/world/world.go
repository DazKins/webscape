package world

import (
	"math/rand"
)

type World struct {
	sizeX int
	sizeY int

	walls [][]bool
}

func NewWorld(sizeX int, sizeY int) *World {
	walls := make([][]bool, sizeX)
	for i := range walls {
		walls[i] = make([]bool, sizeY)
		for j := range walls[i] {
			if rand.Intn(100) < 10 {
				walls[i][j] = true
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

func (w *World) GetWall(x int, y int) bool {
	if x < -w.sizeX/2 || x > w.sizeX/2-1 || y < -w.sizeY/2 || y > w.sizeY/2-1 {
		return true
	}

	return w.walls[x+w.sizeX/2][y+w.sizeY/2]
}

func (w *World) GetWalls() [][]bool {
	return w.walls
}
