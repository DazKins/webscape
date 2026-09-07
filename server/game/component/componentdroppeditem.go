package component

import "webscape/server/game/model"

const ComponentIdDroppedItem = ComponentId("droppeditem")

// CDroppedItem holds the actual item until it is transferred to an inventory.
// Public presentation is replicated through metadata and renderable components.
type CDroppedItem struct {
	Item *model.Item
}

func (c *CDroppedItem) GetId() ComponentId { return ComponentIdDroppedItem }
