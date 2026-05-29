package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"webscape/server/math"
	"webscape/server/util"
)

type World struct {
	sizeX int
	sizeY int

	terrain        []string
	blockers       [][]bool
	wallBlockers   [][]bool
	objectBlockers [][]bool
	walls          []WorldWall
	objects        []WorldObject
	spawns         []WorldSpawn
	conversations  *ConversationRegistry
}

type worldFormat struct {
	FormatVersion int           `json:"formatVersion"`
	Id            string        `json:"id"`
	DisplayName   string        `json:"displayName"`
	Size          worldSize     `json:"size"`
	Terrain       []string      `json:"terrain"`
	Blockers      []bool        `json:"blockers"`
	Walls         []WorldWall   `json:"walls"`
	Objects       []WorldObject `json:"objects"`
	Spawns        []WorldSpawn  `json:"spawns"`
}

type gameFormat struct {
	FormatVersion int             `json:"formatVersion"`
	Id            string          `json:"id"`
	DisplayName   string          `json:"displayName"`
	Files         gameFormatFiles `json:"files"`
}

type gameFormatFiles struct {
	Maps          []string `json:"maps"`
	Conversations []string `json:"conversations"`
	Quests        []string `json:"quests"`
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

type WorldWall struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type WorldSpawn struct {
	Type           string `json:"type"`
	X              int    `json:"x"`
	Y              int    `json:"y"`
	EntityType     string `json:"entityType"`
	Name           string `json:"name"`
	ConversationId string `json:"conversationId"`
}

func LoadFromGameFolder(path string) (*World, error) {
	return LoadFromGameFS(os.DirFS(path))
}

func LoadFromGameFS(gameFS fs.FS) (*World, error) {
	data, err := fs.ReadFile(gameFS, "game.json")
	if err != nil {
		return nil, fmt.Errorf("read game.json: %w", err)
	}

	var format gameFormat
	if err := json.Unmarshal(data, &format); err != nil {
		return nil, fmt.Errorf("parse game.json: %w", err)
	}
	if err := validateGameFormat(format); err != nil {
		return nil, err
	}

	conversations, err := loadConversationRegistry(gameFS, format.Files.Conversations)
	if err != nil {
		return nil, err
	}

	mapPath := format.Files.Maps[0]
	mapData, err := fs.ReadFile(gameFS, mapPath)
	if err != nil {
		return nil, fmt.Errorf("read map %q: %w", mapPath, err)
	}

	world, err := loadWorldFromBytes(mapData)
	if err != nil {
		return nil, err
	}
	world.conversations = conversations
	return world, nil
}

func loadWorldFromBytes(data []byte) (*World, error) {
	var format worldFormat
	if err := json.Unmarshal(data, &format); err != nil {
		return nil, err
	}

	if err := validateWorldFormat(format); err != nil {
		return nil, err
	}

	blockers := makeBlockerGrid(format.Size.X, format.Size.Y, format.Blockers)
	wallBlockers := makeBlockerGrid(format.Size.X, format.Size.Y, nil)
	for _, wall := range format.Walls {
		wallBlockers[wall.X][wall.Y] = true
	}

	objectBlockers := makeBlockerGrid(format.Size.X, format.Size.Y, nil)
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
				objectBlockers[x][y] = true
			}
		}
	}

	return &World{
		sizeX:          format.Size.X,
		sizeY:          format.Size.Y,
		terrain:        format.Terrain,
		blockers:       blockers,
		wallBlockers:   wallBlockers,
		objectBlockers: objectBlockers,
		walls:          format.Walls,
		objects:        format.Objects,
		spawns:         format.Spawns,
	}, nil
}

func NewWorld(sizeX int, sizeY int) *World {
	terrain := make([]string, sizeX*sizeY)
	for i := range terrain {
		terrain[i] = "grass"
	}

	return &World{
		sizeX:          sizeX,
		sizeY:          sizeY,
		terrain:        terrain,
		blockers:       makeBlockerGrid(sizeX, sizeY, nil),
		wallBlockers:   makeBlockerGrid(sizeX, sizeY, nil),
		objectBlockers: makeBlockerGrid(sizeX, sizeY, nil),
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

	return w.blockers[x][y] || w.wallBlockers[x][y] || w.objectBlockers[x][y]
}

func (w *World) GetBlockers() [][]bool {
	return w.blockers
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

func (w *World) GetWalls() []WorldWall {
	result := make([]WorldWall, len(w.walls))
	copy(result, w.walls)
	return result
}

func (w *World) GetSpawns() []WorldSpawn {
	result := make([]WorldSpawn, len(w.spawns))
	copy(result, w.spawns)
	return result
}

func (w *World) GetConversation(id string) (*Conversation, bool) {
	return w.conversations.Get(id)
}

func (w *World) GetConversationRegistry() *ConversationRegistry {
	return w.conversations
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
	if len(format.Blockers) > 0 && len(format.Blockers) != tileCount {
		return fmt.Errorf("blockers length must be %d", tileCount)
	}
	for _, wall := range format.Walls {
		if wall.Id == "" || wall.Type == "" {
			return errors.New("world walls must have id and type")
		}
		if wall.X < 0 || wall.Y < 0 || wall.X >= format.Size.X || wall.Y >= format.Size.Y {
			return fmt.Errorf("wall %q is out of bounds", wall.Id)
		}
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

func validateGameFormat(format gameFormat) error {
	if format.FormatVersion != 1 {
		return fmt.Errorf("unsupported game format version %d", format.FormatVersion)
	}
	if format.Id == "" {
		return errors.New("game id is required")
	}
	if len(format.Files.Maps) == 0 {
		return errors.New("game must include at least one map")
	}
	for _, mapPath := range format.Files.Maps {
		if !fs.ValidPath(mapPath) || mapPath == "." {
			return fmt.Errorf("invalid map path %q", mapPath)
		}
	}
	for _, conversationPath := range format.Files.Conversations {
		if !fs.ValidPath(conversationPath) || conversationPath == "." {
			return fmt.Errorf("invalid conversation path %q", conversationPath)
		}
	}
	for _, questPath := range format.Files.Quests {
		if !fs.ValidPath(questPath) || questPath == "." {
			return fmt.Errorf("invalid quest path %q", questPath)
		}
	}
	return nil
}

func makeBlockerGrid(sizeX int, sizeY int, blockers []bool) [][]bool {
	grid := make([][]bool, sizeX)
	for x := range grid {
		grid[x] = make([]bool, sizeY)
		for y := range grid[x] {
			index := y*sizeX + x
			grid[x][y] = index < len(blockers) && blockers[index]
		}
	}
	return grid
}
