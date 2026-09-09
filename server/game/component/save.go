package component

// Save codecs are a separate contract from client replication. Adding a component
// requires an explicit durable codec or an explicit transient classification here.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

type SavedComponent struct {
	Version int             `json:"version"`
	Data    json.RawMessage `json:"data"`
}

func transient(id ComponentId) bool {
	switch id {
	case ComponentIdPathing, ComponentIdInteracting, ComponentIdCombatState,
		ComponentIdWoodcutting, ComponentIdFishing, ComponentIdFacing,
		ComponentIdActiveConversation, ComponentIdTrading:
		return true
	}
	return false
}

func marshalSaved(value any) (SavedComponent, error) {
	data, err := json.Marshal(value)
	return SavedComponent{Version: 1, Data: data}, err
}

func decodeSaved(data []byte, target any) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("null component payload")
	}
	// Every field in a v1 DTO is required (nullable fields still need a key).
	// This prevents truncated objects from silently becoming valid zero values.
	shape := reflect.TypeOf(target).Elem()
	if shape.Kind() == reflect.Struct {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		for i := 0; i < shape.NumField(); i++ {
			field := shape.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" {
				name = field.Name
			}
			if _, ok := fields[name]; !ok {
				return fmt.Errorf("missing saved field %s", name)
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing component data")
	}
	return nil
}

type savedPosition struct {
	Position math.Vec2 `json:"position"`
}

type savedRenderable struct {
	RenderType  string `json:"renderType"`
	Orientation string `json:"orientation"`
}

type savedHealth struct {
	MaxHealth     int `json:"maxHealth"`
	CurrentHealth int `json:"currentHealth"`
}

type savedAppearance struct {
	Appearance Appearance `json:"appearance"`
}

type savedBaseStats struct {
	Strength  int `json:"strength"`
	Dexterity int `json:"dexterity"`
	Vitality  int `json:"vitality"`
}

type savedCombatStats struct {
	MinDamage        int                `json:"minDamage"`
	MaxDamage        int                `json:"maxDamage"`
	Accuracy         int                `json:"accuracy"`
	Evasion          int                `json:"evasion"`
	Armor            int                `json:"armor"`
	CritChance       float64            `json:"critChance"`
	CritMultiplier   float64            `json:"critMultiplier"`
	AttackRange      int                `json:"attackRange"`
	AttackSpeedTicks int                `json:"attackSpeedTicks"`
	AttackMethod     model.AttackMethod `json:"attackMethod"`
	WindUpTicks      int                `json:"windUpTicks"`
	TravelTicks      int                `json:"travelTicks"`
	ProjectileType   string             `json:"projectileType"`
}

type savedPlayer struct {
	Name string `json:"name"`
}

type savedInventory struct {
	Items []*model.Item `json:"items"`
}

type savedEquipped struct {
	Slots map[model.EquipmentSlot]*model.Item `json:"slots"`
}

type savedConversation struct {
	ConversationId string `json:"conversationId"`
}

type savedOpenable struct {
	IsOpen bool `json:"isOpen"`
}

type savedLootable struct {
	Once   bool       `json:"once"`
	Looted bool       `json:"looted"`
	Items  []LootItem `json:"items"`
}

type savedRandomWalk struct {
	WalkTimer   int       `json:"walkTimer"`
	MaxDistance int       `json:"maxDistance"`
	Origin      math.Vec2 `json:"origin"`
	HasOrigin   bool      `json:"hasOrigin"`
}

type savedFishable struct {
	CatchChancePercent int      `json:"catchChancePercent"`
	Yield              LootItem `json:"yield"`
}

type savedWoodcuttable struct {
	MaxDurability         int            `json:"maxDurability"`
	CurrentDurability     int            `json:"currentDurability"`
	RespawnTicks          int            `json:"respawnTicks"`
	Yield                 LootItem       `json:"yield"`
	Depleted              bool           `json:"depleted"`
	RemainingRespawnTicks int            `json:"remainingRespawnTicks"`
	LastFellerEntityId    model.EntityId `json:"lastFellerEntityId"`
}

type savedShop struct {
	Offers []ShopOffer `json:"offers"`
}

type savedDroppedItem struct {
	Item *model.Item `json:"item"`
}

type savedRewardDrop struct {
}

type savedSpawn struct {
	Position       math.Vec2       `json:"position"`
	RespawnTicks   int             `json:"respawnTicks"`
	RemainingTicks int             `json:"remainingTicks"`
	TemplateID     string          `json:"templateId"`
	Template       map[string]any  `json:"template"`
	ChildID        *model.EntityId `json:"childId"`
	HasSpawned     bool            `json:"hasSpawned"`
}
type savedQuestLog struct {
	Active    map[string]*QuestProgress `json:"active"`
	Completed []CompletedQuest          `json:"completed"`
}
type savedLogEntry struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}
type savedCombatLog struct {
	Entries    []savedLogEntry `json:"entries"`
	MaxEntries int             `json:"maxEntries"`
}

// SaveComponent returns a zero-version record for explicitly transient components.
func SaveComponent(value Component) (SavedComponent, error) {
	if transient(value.GetId()) {
		return SavedComponent{}, nil
	}
	switch c := value.(type) {
	case *CPosition:
		return marshalSaved(savedPosition{Position: c.position})
	case *CRenderable:
		return marshalSaved(savedRenderable{RenderType: c.renderType, Orientation: c.orientation})
	case *CHealth:
		return marshalSaved(savedHealth{MaxHealth: c.maxHealth, CurrentHealth: c.currentHealth})
	case *CAppearance:
		return marshalSaved(savedAppearance{Appearance: c.appearance})
	case *CBaseStats:
		return marshalSaved(savedBaseStats{Strength: c.strength, Dexterity: c.dexterity, Vitality: c.vitality})
	case *CCombatStats:
		return marshalSaved(savedCombatStats{MinDamage: c.minDamage, MaxDamage: c.maxDamage, Accuracy: c.accuracy, Evasion: c.evasion, Armor: c.armor, CritChance: c.critChance, CritMultiplier: c.critMultiplier, AttackRange: c.attackRange, AttackSpeedTicks: c.attackSpeedTicks, AttackMethod: c.attackMethod, WindUpTicks: c.windUpTicks, TravelTicks: c.travelTicks, ProjectileType: c.projectileType})
	case *CPlayer:
		return marshalSaved(savedPlayer{Name: c.name})
	case *CInventory:
		return marshalSaved(savedInventory{Items: c.items})
	case *CEquipped:
		return marshalSaved(savedEquipped{Slots: c.slots})
	case *CConversation:
		return marshalSaved(savedConversation{ConversationId: c.conversationId})
	case *COpenable:
		return marshalSaved(savedOpenable{IsOpen: c.isOpen})
	case *CLootable:
		return marshalSaved(savedLootable{Once: c.once, Looted: c.looted, Items: c.items})
	case *CRandomWalk:
		return marshalSaved(savedRandomWalk{WalkTimer: c.walkTimer, MaxDistance: c.maxDistance, Origin: c.origin, HasOrigin: c.hasOrigin})
	case *CFishable:
		return marshalSaved(savedFishable{CatchChancePercent: c.catchChancePercent, Yield: c.yield})
	case *CWoodcuttable:
		return marshalSaved(savedWoodcuttable{MaxDurability: c.maxDurability, CurrentDurability: c.currentDurability, RespawnTicks: c.respawnTicks, Yield: c.yield, Depleted: c.depleted, RemainingRespawnTicks: c.remainingRespawnTicks, LastFellerEntityId: c.lastFellerEntityId})
	case *CShop:
		return marshalSaved(savedShop{Offers: c.Offers})
	case *CDroppedItem:
		return marshalSaved(savedDroppedItem{Item: c.Item})
	case *CRewardDrop:
		return marshalSaved(savedRewardDrop{})
	case *CMetadata:
		return marshalSaved(c.metadata)
	case *CQuestLog:
		return marshalSaved(savedQuestLog{Active: c.active, Completed: c.completed})
	case *CSpawn:
		s := savedSpawn{Position: c.spawnPosition, RespawnTicks: c.respawnTicks, RemainingTicks: c.remainingRespawnTicks, TemplateID: c.childTemplateEntityId, Template: c.childTemplateComponents, HasSpawned: c.hasSpawned}
		if c.childEntityId.IsPresent() {
			id := c.childEntityId.Unwrap()
			s.ChildID = &id
		}
		return marshalSaved(s)
	case *CCombatLog:
		s := savedCombatLog{MaxEntries: c.maxEntries, Entries: []savedLogEntry{}}
		for _, e := range c.entries {
			s.Entries = append(s.Entries, savedLogEntry{e.text, e.kind})
		}
		return marshalSaved(s)
	case *CLocomotion:
		return marshalSaved(struct{}{})

	default:
		return SavedComponent{}, fmt.Errorf("component %q has no save codec", value.GetId())
	}
}

func RestoreComponent(id ComponentId, saved SavedComponent) (Component, error) {
	if saved.Version != 1 {
		return nil, fmt.Errorf("unsupported component %q version %d", id, saved.Version)
	}
	switch id {
	case ComponentIdPosition:
		var s savedPosition
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CPosition{position: s.Position}, nil
	case ComponentIdRenderable:
		var s savedRenderable
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CRenderable{renderType: s.RenderType, orientation: s.Orientation}, nil
	case ComponentIdHealth:
		var s savedHealth
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CHealth{maxHealth: s.MaxHealth, currentHealth: s.CurrentHealth}, nil
	case ComponentIdAppearance:
		var s savedAppearance
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CAppearance{appearance: s.Appearance}, nil
	case ComponentIdBaseStats:
		var s savedBaseStats
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CBaseStats{strength: s.Strength, dexterity: s.Dexterity, vitality: s.Vitality}, nil
	case ComponentIdCombatStats:
		var s savedCombatStats
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CCombatStats{minDamage: s.MinDamage, maxDamage: s.MaxDamage, accuracy: s.Accuracy, evasion: s.Evasion, armor: s.Armor, critChance: s.CritChance, critMultiplier: s.CritMultiplier, attackRange: s.AttackRange, attackSpeedTicks: s.AttackSpeedTicks, attackMethod: s.AttackMethod, windUpTicks: s.WindUpTicks, travelTicks: s.TravelTicks, projectileType: s.ProjectileType}, nil
	case ComponentIdPlayer:
		var s savedPlayer
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CPlayer{name: s.Name}, nil
	case ComponentIdInventory:
		var s savedInventory
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CInventory{items: s.Items}, nil
	case ComponentIdEquipped:
		var s savedEquipped
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CEquipped{slots: s.Slots}, nil
	case ComponentIdConversation:
		var s savedConversation
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CConversation{conversationId: s.ConversationId}, nil
	case ComponentIdOpenable:
		var s savedOpenable
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &COpenable{isOpen: s.IsOpen}, nil
	case ComponentIdLootable:
		var s savedLootable
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CLootable{once: s.Once, looted: s.Looted, items: s.Items}, nil
	case ComponentIdRandomWalk:
		var s savedRandomWalk
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CRandomWalk{walkTimer: s.WalkTimer, maxDistance: s.MaxDistance, origin: s.Origin, hasOrigin: s.HasOrigin}, nil
	case ComponentIdFishable:
		var s savedFishable
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CFishable{catchChancePercent: s.CatchChancePercent, yield: s.Yield}, nil
	case ComponentIdWoodcuttable:
		var s savedWoodcuttable
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CWoodcuttable{maxDurability: s.MaxDurability, currentDurability: s.CurrentDurability, respawnTicks: s.RespawnTicks, yield: s.Yield, depleted: s.Depleted, remainingRespawnTicks: s.RemainingRespawnTicks, lastFellerEntityId: s.LastFellerEntityId}, nil
	case ComponentIdShop:
		var s savedShop
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CShop{Offers: s.Offers}, nil
	case ComponentIdDroppedItem:
		var s savedDroppedItem
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CDroppedItem{Item: s.Item}, nil
	case ComponentIdRewardDrop:
		var s savedRewardDrop
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return &CRewardDrop{}, nil
	case ComponentIdMetadata:
		var raw any
		if err := decodeSaved(saved.Data, &raw); err != nil {
			return nil, err
		}
		obj, ok := raw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("metadata must be an object")
		}
		return NewCMetadata(saveJSON(obj)), nil
	case ComponentIdQuestLog:
		var s savedQuestLog
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		c := NewCQuestLog()
		for id, p := range s.Active {
			if p == nil || p.QuestId != id || p.CurrentStepIndex < 0 || p.CurrentCount < 0 {
				return nil, fmt.Errorf("invalid quest progress")
			}
			c.SetProgress(id, p.CurrentStepIndex, p.StepId, p.CurrentCount)
		}
		for _, q := range s.Completed {
			c.CompleteQuest(q.QuestId)
		}
		return c, nil
	case ComponentIdSpawn:
		var s savedSpawn
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		c := NewCSpawn(s.Position, s.RespawnTicks, s.TemplateID, s.Template)
		c.remainingRespawnTicks, c.hasSpawned = s.RemainingTicks, s.HasSpawned
		if s.ChildID != nil {
			c.SetChildEntityId(*s.ChildID)
		}
		return c, nil
	case ComponentIdCombatLog:
		var s savedCombatLog
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		if s.MaxEntries < 0 || len(s.Entries) > s.MaxEntries {
			return nil, fmt.Errorf("invalid combat log capacity")
		}
		c := NewCCombatLog(s.MaxEntries)
		for _, e := range s.Entries {
			c.AddEntry(NewCombatLogEntry(e.Text, e.Kind))
		}
		return c, nil
	case ComponentIdLocomotion:
		var s struct{}
		if err := decodeSaved(saved.Data, &s); err != nil {
			return nil, err
		}
		return NewCLocomotion(LocomotionPhaseIdle, 0), nil

	default:
		return nil, fmt.Errorf("unknown durable component %q", id)
	}
}

// saveJSON converts decoded metadata back to the game's JSON value types.
func saveJSON(v any) util.Json {
	switch v := v.(type) {
	case nil:
		return util.JNull{}
	case bool:
		return util.JBool(v)
	case float64:
		return util.JNumber(v)
	case string:
		return util.JString(v)
	case []any:
		result := make(util.JArray, len(v))
		for i, item := range v {
			result[i] = saveJSON(item)
		}
		return result
	case map[string]any:
		result := util.JObject{}
		for key, item := range v {
			result[key] = saveJSON(item)
		}
		return result
	default:
		panic("unsupported decoded JSON type")
	}
}
