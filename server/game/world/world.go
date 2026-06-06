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
	entities       []WorldEntity
	conversations  *ConversationRegistry
	quests         *QuestRegistry
}

type worldFormat struct {
	FormatVersion int           `json:"formatVersion"`
	Id            string        `json:"id"`
	DisplayName   string        `json:"displayName"`
	Size          worldSize     `json:"size"`
	Terrain       []string      `json:"terrain"`
	Blockers      []bool        `json:"blockers"`
	Walls         []WorldWall   `json:"walls"`
	Entities      []WorldEntity `json:"entities"`
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

type WorldEntity struct {
	Id         string         `json:"id"`
	Components map[string]any `json:"components"`
}

type WorldWall struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
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
	quests, err := loadQuestRegistry(gameFS, format.Files.Quests)
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
	world.quests = quests
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

	entities := make([]WorldEntity, 0, len(format.Entities))
	entities = append(entities, format.Entities...)

	objectBlockers := makeBlockerGrid(format.Size.X, format.Size.Y, nil)
	for _, entity := range entities {
		position, ok := entityPosition(entity)
		if !ok {
			continue
		}
		width, height := entitySize(entity)
		if !entityBlocksMovement(entity) {
			continue
		}
		for x := position.X; x < position.X+width; x++ {
			for y := position.Y; y < position.Y+height; y++ {
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
		entities:       entities,
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

func (w *World) GetEntities() []WorldEntity {
	result := make([]WorldEntity, len(w.entities))
	copy(result, w.entities)
	return result
}

func (w *World) GetWalls() []WorldWall {
	result := make([]WorldWall, len(w.walls))
	copy(result, w.walls)
	return result
}

func (w *World) GetConversation(id string) (*Conversation, bool) {
	return w.conversations.Get(id)
}

func (w *World) GetConversationRegistry() *ConversationRegistry {
	return w.conversations
}

func (w *World) GetQuest(id string) (*Quest, bool) {
	return w.quests.Get(id)
}

func (w *World) GetQuestRegistry() *QuestRegistry {
	return w.quests
}

func (w *World) GetPlayerSpawn() math.Vec2 {
	for _, entity := range w.entities {
		if !hasComponent(entity, "playerSpawn") {
			continue
		}
		position, ok := entityPosition(entity)
		if ok {
			return position
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
	for _, entity := range format.Entities {
		if entity.Id == "" {
			return errors.New("world entities must have id")
		}
		position, ok := entityPosition(entity)
		if !ok {
			return fmt.Errorf("entity %q must include a position component", entity.Id)
		}
		width, height := entitySize(entity)
		if position.X < 0 || position.Y < 0 || position.X+width > format.Size.X || position.Y+height > format.Size.Y {
			return fmt.Errorf("entity %q is out of bounds", entity.Id)
		}
		if spawnTemplateComponents, ok := entitySpawnTemplateComponents(entity); ok {
			if _, hasPosition := spawnTemplateComponents["position"]; hasPosition {
				return fmt.Errorf("spawn entity %q child template must not include a position component", entity.Id)
			}
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

func hasComponent(entity WorldEntity, componentId string) bool {
	if entity.Components == nil {
		return false
	}
	_, ok := entity.Components[componentId]
	return ok
}

func entityPosition(entity WorldEntity) (math.Vec2, bool) {
	component, ok := entity.Components["position"].(map[string]any)
	if !ok {
		return math.Vec2Zero(), false
	}
	x, okX := numberToInt(component["x"])
	y, okY := numberToInt(component["y"])
	if !okX || !okY {
		return math.Vec2Zero(), false
	}
	return math.Vec2{X: x, Y: y}, true
}

func entitySize(entity WorldEntity) (int, int) {
	metadata, ok := entity.Components["metadata"].(map[string]any)
	if !ok {
		return 1, 1
	}
	width, ok := numberToInt(metadata["width"])
	if !ok || width < 1 {
		width = 1
	}
	height, ok := numberToInt(metadata["height"])
	if !ok || height < 1 {
		height = 1
	}
	return width, height
}

func entityBlocksMovement(entity WorldEntity) bool {
	metadata, ok := entity.Components["metadata"].(map[string]any)
	if !ok {
		return false
	}
	blocksMovement, _ := metadata["blocksMovement"].(bool)
	return blocksMovement
}

func entitySpawnTemplateComponents(entity WorldEntity) (map[string]any, bool) {
	spawn, ok := entity.Components["spawn"].(map[string]any)
	if !ok {
		return nil, false
	}
	rawTemplate, ok := spawn["entity"].(map[string]any)
	if !ok {
		return nil, false
	}
	templateComponents, ok := rawTemplate["components"].(map[string]any)
	return templateComponents, ok
}

func numberToInt(value any) (int, bool) {
	switch value := value.(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	default:
		return 0, false
	}
}
