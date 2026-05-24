package world

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"webscape/server/math"
	"webscape/server/util"
)

//go:embed maps/starter.json
var worldFiles embed.FS

type World struct {
	sizeX int
	sizeY int

	terrain  []string
	walls    [][]bool
	blockers [][]bool
	objects  []WorldObject
	spawns   []WorldSpawn
}

type worldFormat struct {
	FormatVersion int           `json:"formatVersion"`
	Id            string        `json:"id"`
	DisplayName   string        `json:"displayName"`
	Size          worldSize     `json:"size"`
	Terrain       []string      `json:"terrain"`
	Collision     []bool        `json:"collision"`
	Objects       []WorldObject `json:"objects"`
	Spawns        []WorldSpawn  `json:"spawns"`
}

type worldSize struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type WorldObject struct {
	Id             string         `json:"id"`
	Type           string         `json:"type"`
	X              int            `json:"x"`
	Y              int            `json:"y"`
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	BlocksMovement bool           `json:"blocksMovement"`
	Interactable   []string       `json:"interactable"`
	State          map[string]any `json:"state"`
}

type WorldSpawn struct {
	Type       string `json:"type"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	EntityType string `json:"entityType"`
	Name       string `json:"name"`
}

func LoadStarterWorld() (*World, error) {
	return loadWorldFromEmbeddedFile("maps/starter.json")
}

func loadWorldFromEmbeddedFile(path string) (*World, error) {
	data, err := worldFiles.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var format worldFormat
	if err := json.Unmarshal(data, &format); err != nil {
		return nil, err
	}

	if err := validateWorldFormat(format); err != nil {
		return nil, err
	}

	walls := makeWalls(format.Size.X, format.Size.Y, format.Collision)
	blockers := makeWalls(format.Size.X, format.Size.Y, nil)
	for _, object := range format.Objects {
		if !object.BlocksMovement {
			continue
		}
		width := object.Width
		if width == 0 {
			width = 1
		}
		height := object.Height
		if height == 0 {
			height = 1
		}
		for x := object.X; x < object.X+width; x++ {
			for y := object.Y; y < object.Y+height; y++ {
				blockers[x][y] = true
			}
		}
	}

	return &World{
		sizeX:    format.Size.X,
		sizeY:    format.Size.Y,
		terrain:  format.Terrain,
		walls:    walls,
		blockers: blockers,
		objects:  format.Objects,
		spawns:   format.Spawns,
	}, nil
}

func NewWorld(sizeX int, sizeY int) *World {
	walls := make([][]bool, sizeX)
	for i := range walls {
		walls[i] = make([]bool, sizeY)
	}
	terrain := make([]string, sizeX*sizeY)
	for i := range terrain {
		terrain[i] = "grass"
	}

	return &World{
		sizeX:    sizeX,
		sizeY:    sizeY,
		terrain:  terrain,
		walls:    walls,
		blockers: makeWalls(sizeX, sizeY, nil),
	}
}

func (w *World) GetSizeX() int {
	return w.sizeX
}

func (w *World) GetSizeY() int {
	return w.sizeY
}

func (w *World) GetWall(x int, y int) bool {
	if x < 0 || x >= w.sizeX || y < 0 || y >= w.sizeY {
		return true
	}

	return w.walls[x][y] || w.blockers[x][y]
}

func (w *World) GetWalls() [][]bool {
	return w.walls
}

func (w *World) GetTerrain() []string {
	result := make([]string, len(w.terrain))
	copy(result, w.terrain)
	return result
}

func (w *World) GetObjects() []WorldObject {
	result := make([]WorldObject, len(w.objects))
	copy(result, w.objects)
	return result
}

func (w *World) GetSpawns() []WorldSpawn {
	result := make([]WorldSpawn, len(w.spawns))
	copy(result, w.spawns)
	return result
}

func (w *World) GetPlayerSpawn() math.Vec2 {
	for _, spawn := range w.spawns {
		if spawn.Type == "player" {
			return math.Vec2{X: spawn.X, Y: spawn.Y}
		}
	}
	return math.Vec2Zero()
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

func validateWorldFormat(format worldFormat) error {
	if format.FormatVersion != 1 {
		return fmt.Errorf("unsupported world format version %d", format.FormatVersion)
	}
	if format.Size.X < 1 || format.Size.Y < 1 {
		return errors.New("world size must be positive")
	}
	tileCount := format.Size.X * format.Size.Y
	if len(format.Terrain) != tileCount {
		return fmt.Errorf("terrain length must be %d", tileCount)
	}
	if len(format.Collision) > 0 && len(format.Collision) != tileCount {
		return fmt.Errorf("collision length must be %d", tileCount)
	}
	for _, object := range format.Objects {
		width := object.Width
		if width == 0 {
			width = 1
		}
		height := object.Height
		if height == 0 {
			height = 1
		}
		if object.Id == "" || object.Type == "" {
			return errors.New("world objects must have id and type")
		}
		if object.X < 0 || object.Y < 0 || object.X+width > format.Size.X || object.Y+height > format.Size.Y {
			return fmt.Errorf("object %q is out of bounds", object.Id)
		}
	}
	for _, spawn := range format.Spawns {
		if spawn.Type == "" {
			return errors.New("world spawns must have type")
		}
		if spawn.X < 0 || spawn.X >= format.Size.X || spawn.Y < 0 || spawn.Y >= format.Size.Y {
			return fmt.Errorf("spawn %q at (%d, %d) is out of bounds", spawn.Type, spawn.X, spawn.Y)
		}
	}
	return nil
}

func makeWalls(sizeX int, sizeY int, collision []bool) [][]bool {
	walls := make([][]bool, sizeX)
	for x := range walls {
		walls[x] = make([]bool, sizeY)
		for y := range walls[x] {
			index := y*sizeX + x
			walls[x][y] = index < len(collision) && collision[index]
		}
	}
	return walls
}
