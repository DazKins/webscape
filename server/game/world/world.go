package world

import (
	"errors"
	"log"
	"math/rand"
	"webscape/server/math"
	"webscape/server/util"
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
			if rand.Intn(100) < 30 {
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

func (w *World) GetPath(from math.Vec2, to math.Vec2) (util.Path, error) {
	frontier := util.NewQueue[math.Vec2]()
	cameFrom := make(map[math.Vec2]math.Vec2)
	costSoFar := make(map[math.Vec2]float64)

	frontier.Enqueue(from)
	costSoFar[from] = 0.0

	for frontier.Size() > 0 {
		current := frontier.Dequeue()

		if current.X == to.X && current.Y == to.Y {
			break
		}

		canEast := !w.GetWall(current.X+1, current.Y)
		canWest := !w.GetWall(current.X-1, current.Y)
		canNorth := !w.GetWall(current.X, current.Y+1)
		canSouth := !w.GetWall(current.X, current.Y-1)

		checkStepXMin := -1
		checkStepXMax := 1
		checkStepYMin := -1
		checkStepYMax := 1
		if !canEast {
			checkStepXMax = 0
		}
		if !canWest {
			checkStepXMin = 0
		}
		if !canNorth {
			checkStepYMax = 0
		}
		if !canSouth {
			checkStepYMin = 0
		}

		for x := checkStepXMin; x <= checkStepXMax; x++ {
			for y := checkStepYMin; y <= checkStepYMax; y++ {
				if x == 0 && y == 0 {
					continue
				}

				posX := current.X + x
				posY := current.Y + y

				if w.GetWall(posX, posY) {
					continue
				}

				neighbor := math.Vec2{X: posX, Y: posY}
				dist := 1.0
				if x != 0 && y != 0 {
					dist = 1.0001 // slightly prefer straight line movement
				}
				newCost := costSoFar[current] + dist
				if _, ok := costSoFar[neighbor]; !ok || newCost < costSoFar[neighbor] {
					costSoFar[neighbor] = newCost
					frontier.Enqueue(neighbor)
					cameFrom[neighbor] = current
				}
			}
		}
	}

	path := util.Path{}
	current := to
	for current != from {
		path.Append(current)
		cameFrom, ok := cameFrom[current]
		if !ok {
			log.Println("no path found")
			return util.Path{}, errors.New("no path found")
		}
		current = cameFrom
	}
	return path.Reversed(), nil
}
