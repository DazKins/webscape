package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
	"webscape/server/game/component"
	"webscape/server/math"
)

type ChunkCoord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Chunk struct {
	Id         string
	Coordinate ChunkCoord
	Terrain    []string
	Heights    []int
	Blockers   []bool
	Walls      []WorldWall
}

type World struct {
	chunkSize     ChunkCoord
	chunks        map[ChunkCoord]*Chunk
	entities      []WorldEntity
	playerSpawn   math.Vec2
	conversations *ConversationRegistry
	quests        *QuestRegistry
}

type gameFormat struct {
	FormatVersion int             `json:"formatVersion"`
	Id            string          `json:"id"`
	DisplayName   string          `json:"displayName"`
	World         gameWorldFormat `json:"world"`
	Files         gameFormatFiles `json:"files"`
}

type gameWorldFormat struct {
	ChunkSize ChunkCoord `json:"chunkSize"`
}

type gameFormatFiles struct {
	Chunks        []string `json:"chunks"`
	Conversations []string `json:"conversations"`
	Quests        []string `json:"quests"`
}

type chunkFormat struct {
	FormatVersion int               `json:"formatVersion"`
	Id            string            `json:"id"`
	DisplayName   string            `json:"displayName"`
	Coordinate    ChunkCoord        `json:"coordinate"`
	Terrain       []string          `json:"terrain"`
	Heights       []json.RawMessage `json:"heights"`
	Blockers      []bool            `json:"blockers"`
	Walls         []WorldWall       `json:"walls"`
	Entities      []WorldEntity     `json:"entities"`
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

func LoadFromGameFolder(path string) (*World, error) { return LoadFromGameFS(os.DirFS(path)) }

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

	w := &World{
		chunkSize:     format.World.ChunkSize,
		chunks:        make(map[ChunkCoord]*Chunk),
		conversations: conversations,
		quests:        quests,
	}
	chunkIDs := make(map[string]string)
	entityIDs := make(map[string]string)
	spawnCount := 0
	for _, chunkPath := range format.Files.Chunks {
		chunkData, err := fs.ReadFile(gameFS, chunkPath)
		if err != nil {
			return nil, fmt.Errorf("read chunk %q: %w", chunkPath, err)
		}
		var authored chunkFormat
		if err := json.Unmarshal(chunkData, &authored); err != nil {
			return nil, fmt.Errorf("parse chunk %q: %w", chunkPath, err)
		}
		if err := validateChunkFormat(authored, w.chunkSize); err != nil {
			return nil, fmt.Errorf("chunk %q: %w", chunkPath, err)
		}
		if previous, ok := chunkIDs[authored.Id]; ok {
			return nil, fmt.Errorf("duplicate chunk id %q in %q and %q", authored.Id, previous, chunkPath)
		}
		if _, ok := w.chunks[authored.Coordinate]; ok {
			return nil, fmt.Errorf("duplicate chunk coordinate (%d,%d)", authored.Coordinate.X, authored.Coordinate.Y)
		}
		chunkIDs[authored.Id] = chunkPath
		heights, _ := parseTerrainHeights(authored.Heights, w.chunkSize.X*w.chunkSize.Y)
		w.chunks[authored.Coordinate] = &Chunk{
			Id: authored.Id, Coordinate: authored.Coordinate,
			Terrain: append([]string(nil), authored.Terrain...),
			Heights: heights, Blockers: append([]bool(nil), authored.Blockers...),
			Walls: append([]WorldWall(nil), authored.Walls...),
		}
		origin := w.ChunkOrigin(authored.Coordinate)
		for _, authoredEntity := range authored.Entities {
			if previous, ok := entityIDs[authoredEntity.Id]; ok {
				return nil, fmt.Errorf("duplicate entity id %q in %q and %q", authoredEntity.Id, previous, chunkPath)
			}
			entityIDs[authoredEntity.Id] = chunkPath
			position, _ := entityPosition(authoredEntity)
			global := position.Add(origin)
			positionComponent := authoredEntity.Components["position"].(map[string]any)
			positionComponent["x"] = global.X
			positionComponent["y"] = global.Y
			if hasComponent(authoredEntity, "playerSpawn") {
				spawnCount++
				w.playerSpawn = global
			}
			w.entities = append(w.entities, authoredEntity)
		}
	}
	if spawnCount != 1 {
		return nil, fmt.Errorf("world must contain exactly one playerSpawn, got %d", spawnCount)
	}
	return w, nil
}

// NewWorld remains a compact test constructor. It creates one chunk at (0,0).
func NewWorld(sizeX int, sizeY int) *World {
	tileCount := sizeX * sizeY
	terrain := make([]string, tileCount)
	for i := range terrain {
		terrain[i] = "grass"
	}
	registry := NewConversationRegistry()
	quests := NewQuestRegistry()
	coord := ChunkCoord{}
	return &World{
		chunkSize:     ChunkCoord{X: sizeX, Y: sizeY},
		chunks:        map[ChunkCoord]*Chunk{coord: {Id: "test", Coordinate: coord, Terrain: terrain, Heights: make([]int, tileCount), Blockers: make([]bool, tileCount)}},
		conversations: registry, quests: quests,
	}
}

func (w *World) GetChunkSize() ChunkCoord { return w.chunkSize }
func (w *World) GetSizeX() int            { return w.chunkSize.X }
func (w *World) GetSizeY() int            { return w.chunkSize.Y }

func (w *World) ChunkOrigin(coord ChunkCoord) math.Vec2 {
	return math.Vec2{X: coord.X * w.chunkSize.X, Y: coord.Y * w.chunkSize.Y}
}

func floorDiv(value, divisor int) int {
	quotient := value / divisor
	remainder := value % divisor
	if remainder != 0 && ((remainder < 0) != (divisor < 0)) {
		quotient--
	}
	return quotient
}

func (w *World) GlobalToChunk(x, y int) (ChunkCoord, math.Vec2) {
	coord := ChunkCoord{X: floorDiv(x, w.chunkSize.X), Y: floorDiv(y, w.chunkSize.Y)}
	origin := w.ChunkOrigin(coord)
	return coord, math.Vec2{X: x - origin.X, Y: y - origin.Y}
}

func (w *World) GetChunk(coord ChunkCoord) (*Chunk, bool) {
	chunk, ok := w.chunks[coord]
	return chunk, ok
}

func (w *World) HasChunk(coord ChunkCoord) bool { _, ok := w.chunks[coord]; return ok }

func (w *World) GetChunks() []Chunk {
	coords := make([]ChunkCoord, 0, len(w.chunks))
	for coord := range w.chunks {
		coords = append(coords, coord)
	}
	sort.Slice(coords, func(i, j int) bool {
		if coords[i].Y == coords[j].Y {
			return coords[i].X < coords[j].X
		}
		return coords[i].Y < coords[j].Y
	})
	result := make([]Chunk, 0, len(coords))
	for _, coord := range coords {
		chunk := *w.chunks[coord]
		chunk.Terrain = append([]string(nil), chunk.Terrain...)
		chunk.Heights = append([]int(nil), chunk.Heights...)
		chunk.Blockers = append([]bool(nil), chunk.Blockers...)
		chunk.Walls = append([]WorldWall(nil), chunk.Walls...)
		result = append(result, chunk)
	}
	return result
}

func (w *World) ChunksWithin(center ChunkCoord, radius int) map[ChunkCoord]bool {
	result := make(map[ChunkCoord]bool)
	for y := center.Y - radius; y <= center.Y+radius; y++ {
		for x := center.X - radius; x <= center.X+radius; x++ {
			coord := ChunkCoord{X: x, Y: y}
			if w.HasChunk(coord) {
				result[coord] = true
			}
		}
	}
	return result
}

func (w *World) GetTerrainAt(x, y int) (string, bool) {
	coord, local := w.GlobalToChunk(x, y)
	chunk, ok := w.chunks[coord]
	if !ok {
		return "", false
	}
	return chunk.Terrain[local.Y*w.chunkSize.X+local.X], true
}

func (w *World) GetHeightAt(x, y int) (int, bool) {
	coord, local := w.GlobalToChunk(x, y)
	chunk, ok := w.chunks[coord]
	if !ok {
		return 0, false
	}
	return chunk.Heights[local.Y*w.chunkSize.X+local.X], true
}

func (w *World) GetWall(x, y int) bool { return w.GetStaticWall(x, y) }
func (w *World) GetStaticWall(x, y int) bool {
	coord, local := w.GlobalToChunk(x, y)
	chunk, ok := w.chunks[coord]
	if !ok {
		return true
	}
	index := local.Y*w.chunkSize.X + local.X
	if chunk.Blockers[index] {
		return true
	}
	for _, wall := range chunk.Walls {
		if wall.X == local.X && wall.Y == local.Y {
			return true
		}
	}
	return false
}

// Compatibility accessors expose the (0,0) chunk without exposing all chunks to clients.
func (w *World) GetTerrain() []string {
	if c, ok := w.chunks[ChunkCoord{}]; ok {
		return append([]string(nil), c.Terrain...)
	}
	return nil
}
func (w *World) GetHeights() []int {
	if c, ok := w.chunks[ChunkCoord{}]; ok {
		return append([]int(nil), c.Heights...)
	}
	return nil
}
func (w *World) GetBlockers() [][]bool {
	grid := make([][]bool, w.chunkSize.X)
	for x := range grid {
		grid[x] = make([]bool, w.chunkSize.Y)
	}
	if c, ok := w.chunks[ChunkCoord{}]; ok {
		for y := 0; y < w.chunkSize.Y; y++ {
			for x := 0; x < w.chunkSize.X; x++ {
				grid[x][y] = c.Blockers[y*w.chunkSize.X+x]
			}
		}
	}
	return grid
}
func (w *World) GetWalls() []WorldWall {
	if c, ok := w.chunks[ChunkCoord{}]; ok {
		return append([]WorldWall(nil), c.Walls...)
	}
	return nil
}
func (w *World) GetEntities() []WorldEntity                      { return append([]WorldEntity(nil), w.entities...) }
func (w *World) GetPlayerSpawn() math.Vec2                       { return w.playerSpawn }
func (w *World) GetConversation(id string) (*Conversation, bool) { return w.conversations.Get(id) }
func (w *World) GetConversationRegistry() *ConversationRegistry  { return w.conversations }
func (w *World) GetQuest(id string) (*Quest, bool)               { return w.quests.Get(id) }
func (w *World) GetQuestRegistry() *QuestRegistry                { return w.quests }

func validateGameFormat(format gameFormat) error {
	if format.FormatVersion != 2 {
		return fmt.Errorf("unsupported game format version %d", format.FormatVersion)
	}
	if format.Id == "" {
		return errors.New("game id is required")
	}
	if format.World.ChunkSize.X < 1 || format.World.ChunkSize.Y < 1 {
		return errors.New("world.chunkSize must be positive")
	}
	if len(format.Files.Chunks) == 0 {
		return errors.New("game must include at least one chunk")
	}
	seen := make(map[string]bool)
	validatePaths := func(label string, paths []string) error {
		for _, path := range paths {
			if !fs.ValidPath(path) || path == "." || strings.Contains(path, `\`) {
				return fmt.Errorf("invalid %s path %q", label, path)
			}
			if seen[path] {
				return fmt.Errorf("duplicate project file path %q", path)
			}
			seen[path] = true
		}
		return nil
	}
	if err := validatePaths("chunk", format.Files.Chunks); err != nil {
		return err
	}
	if err := validatePaths("conversation", format.Files.Conversations); err != nil {
		return err
	}
	return validatePaths("quest", format.Files.Quests)
}

func validateChunkFormat(format chunkFormat, size ChunkCoord) error {
	if format.FormatVersion != 2 {
		return fmt.Errorf("unsupported chunk format version %d", format.FormatVersion)
	}
	if format.Id == "" {
		return errors.New("chunk id is required")
	}
	tileCount := size.X * size.Y
	if len(format.Terrain) != tileCount {
		return fmt.Errorf("terrain length must be %d", tileCount)
	}
	if _, err := parseTerrainHeights(format.Heights, tileCount); err != nil {
		return err
	}
	if len(format.Blockers) != tileCount {
		return fmt.Errorf("blockers length must be %d", tileCount)
	}
	wallIDs := make(map[string]bool)
	for _, wall := range format.Walls {
		if wall.Id == "" || wall.Type == "" {
			return errors.New("chunk walls must have id and type")
		}
		if wallIDs[wall.Id] {
			return fmt.Errorf("duplicate wall id %q", wall.Id)
		}
		wallIDs[wall.Id] = true
		if wall.X < 0 || wall.Y < 0 || wall.X >= size.X || wall.Y >= size.Y {
			return fmt.Errorf("wall %q is out of bounds", wall.Id)
		}
	}
	entityIDs := make(map[string]bool)
	for _, entity := range format.Entities {
		if entity.Id == "" {
			return errors.New("chunk entities must have id")
		}
		if entityIDs[entity.Id] {
			return fmt.Errorf("duplicate entity id %q", entity.Id)
		}
		entityIDs[entity.Id] = true
		position, ok := entityPosition(entity)
		if !ok {
			return fmt.Errorf("entity %q must include a position component", entity.Id)
		}
		width, height := entitySize(entity)
		if position.X < 0 || position.Y < 0 || position.X+width > size.X || position.Y+height > size.Y {
			return fmt.Errorf("entity %q footprint is out of chunk bounds", entity.Id)
		}
		if template, ok := entitySpawnTemplateComponents(entity); ok {
			if _, hasPosition := template["position"]; hasPosition {
				return fmt.Errorf("spawn entity %q child template must not include a position component", entity.Id)
			}
			if err := validateWoodcuttableComponent(entity.Id+" child template", template); err != nil {
				return err
			}
			if err := validateFishableComponent(entity.Id+" child template", template); err != nil {
				return err
			}
			if err := validateAppearanceComponent(entity.Id+" child template", template); err != nil {
				return err
			}
		}
		if err := validateWoodcuttableComponent(entity.Id, entity.Components); err != nil {
			return err
		}
		if err := validateFishableComponent(entity.Id, entity.Components); err != nil {
			return err
		}
		if err := validateAppearanceComponent(entity.Id, entity.Components); err != nil {
			return err
		}
	}
	return nil
}

func validateFishableComponent(entityId string, components map[string]any) error {
	rawValue, exists := components["fishable"]
	if !exists {
		return nil
	}
	raw, ok := rawValue.(map[string]any)
	if !ok {
		return fmt.Errorf("entity %q fishable must be an object", entityId)
	}
	catchChancePercent, ok := numberToInt(raw["catchChancePercent"])
	if !ok || catchChancePercent < 1 || catchChancePercent > 100 {
		return fmt.Errorf("entity %q fishable.catchChancePercent must be an integer from 1 to 100", entityId)
	}
	yield, ok := raw["yield"].(map[string]any)
	if !ok {
		return fmt.Errorf("entity %q fishable.yield must be an object", entityId)
	}
	name, nameOK := yield["name"].(string)
	if !nameOK || strings.TrimSpace(name) == "" {
		return fmt.Errorf("entity %q fishable.yield.name must be a non-empty string", entityId)
	}
	itemType, typeOK := yield["type"].(string)
	if !typeOK || strings.TrimSpace(itemType) == "" {
		return fmt.Errorf("entity %q fishable.yield.type must be a non-empty string", entityId)
	}
	count, ok := numberToInt(yield["count"])
	if !ok || count < 1 {
		return fmt.Errorf("entity %q fishable.yield.count must be a positive integer", entityId)
	}
	return nil
}

func validateAppearanceComponent(entityID string, components map[string]any) error {
	raw, exists := components["appearance"]
	if !exists {
		return nil
	}
	if _, err := component.ParseAppearance(raw); err != nil {
		return fmt.Errorf("entity %q %w", entityID, err)
	}
	return nil
}

func validateWoodcuttableComponent(entityId string, components map[string]any) error {
	rawValue, exists := components["woodcuttable"]
	if !exists {
		return nil
	}
	raw, ok := rawValue.(map[string]any)
	if !ok {
		return fmt.Errorf("entity %q woodcuttable must be an object", entityId)
	}
	maxDurability, ok := numberToInt(raw["maxDurability"])
	if !ok || maxDurability < 1 {
		return fmt.Errorf("entity %q woodcuttable.maxDurability must be a positive integer", entityId)
	}
	respawnTicks, ok := numberToInt(raw["respawnTicks"])
	if !ok || respawnTicks < 1 {
		return fmt.Errorf("entity %q woodcuttable.respawnTicks must be a positive integer", entityId)
	}
	yield, ok := raw["yield"].(map[string]any)
	if !ok {
		return fmt.Errorf("entity %q woodcuttable.yield must be an object", entityId)
	}
	name, nameOK := yield["name"].(string)
	if !nameOK || strings.TrimSpace(name) == "" {
		return fmt.Errorf("entity %q woodcuttable.yield.name must be a non-empty string", entityId)
	}
	itemType, typeOK := yield["type"].(string)
	if !typeOK || itemType != "material" {
		return fmt.Errorf("entity %q woodcuttable.yield.type must be %q", entityId, "material")
	}
	count, ok := numberToInt(yield["count"])
	if !ok || count < 1 {
		return fmt.Errorf("entity %q woodcuttable.yield.count must be a positive integer", entityId)
	}
	return nil
}

func parseTerrainHeights(raw []json.RawMessage, count int) ([]int, error) {
	if len(raw) != count {
		return nil, fmt.Errorf("heights length must be %d", count)
	}
	result := make([]int, len(raw))
	for i, value := range raw {
		if err := json.Unmarshal(value, &result[i]); err != nil || result[i] < 0 || result[i] > 10 {
			return nil, fmt.Errorf("heights[%d] must be an integer from 0 to 10", i)
		}
	}
	return result, nil
}

func hasComponent(entity WorldEntity, id string) bool { _, ok := entity.Components[id]; return ok }
func entityPosition(entity WorldEntity) (math.Vec2, bool) {
	raw, ok := entity.Components["position"].(map[string]any)
	if !ok {
		return math.Vec2Zero(), false
	}
	x, okX := numberToInt(raw["x"])
	y, okY := numberToInt(raw["y"])
	return math.Vec2{X: x, Y: y}, okX && okY
}
func entitySize(entity WorldEntity) (int, int) {
	metadata, ok := entity.Components["metadata"].(map[string]any)
	if !ok {
		return 1, 1
	}
	width, okW := numberToInt(metadata["width"])
	if !okW || width < 1 {
		width = 1
	}
	height, okH := numberToInt(metadata["height"])
	if !okH || height < 1 {
		height = 1
	}
	return width, height
}
func entitySpawnTemplateComponents(entity WorldEntity) (map[string]any, bool) {
	spawn, ok := entity.Components["spawn"].(map[string]any)
	if !ok {
		return nil, false
	}
	template, ok := spawn["entity"].(map[string]any)
	if !ok {
		return nil, false
	}
	components, ok := template["components"].(map[string]any)
	return components, ok
}
func numberToInt(value any) (int, bool) {
	switch value := value.(type) {
	case float64:
		if value != float64(int(value)) {
			return 0, false
		}
		return int(value), true
	case int:
		return value, true
	default:
		return 0, false
	}
}
