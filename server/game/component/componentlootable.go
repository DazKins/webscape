package component

import (
	"webscape/server/game/model"
	"webscape/server/util"
)

const ComponentIdLootable = ComponentId("lootable")

type LootItem struct {
	Name  string
	Type  string
	Count int
}

type CLootable struct {
	once   bool
	looted bool
	items  []LootItem
}

func NewCLootable(once bool, items []LootItem) *CLootable {
	copiedItems := make([]LootItem, len(items))
	copy(copiedItems, items)
	return &CLootable{
		once:  once,
		items: copiedItems,
	}
}

func (c *CLootable) GetId() ComponentId {
	return ComponentIdLootable
}

func (c *CLootable) Serialize() util.Json {
	return util.JObject(map[string]util.Json{
		"once":      util.JBool(c.once),
		"looted":    util.JBool(c.looted),
		"itemCount": util.JNumber(c.ItemCount()),
	})
}

func (c *CLootable) IsOnce() bool {
	return c.once
}

func (c *CLootable) IsLooted() bool {
	return c.looted
}

func (c *CLootable) SetLooted(looted bool) {
	c.looted = looted
}

func (c *CLootable) CanLoot() bool {
	return !c.once || !c.looted
}

func (c *CLootable) GetItems() []LootItem {
	result := make([]LootItem, len(c.items))
	copy(result, c.items)
	return result
}

func (c *CLootable) ItemCount() int {
	count := 0
	for _, item := range c.items {
		if item.Count < 1 {
			count += 1
			continue
		}
		count += item.Count
	}
	return count
}

func (item LootItem) CreateItem() *model.Item {
	return model.NewItem(item.Name, item.Type)
}
