package component

const ComponentIdFishable = ComponentId("fishable")

type CFishable struct {
	catchChancePercent int
	yield              LootItem
}

func NewCFishable(catchChancePercent int, yield LootItem) *CFishable {
	return &CFishable{
		catchChancePercent: catchChancePercent,
		yield:              yield,
	}
}

func (c *CFishable) GetId() ComponentId {
	return ComponentIdFishable
}

func (c *CFishable) GetCatchChancePercent() int {
	return c.catchChancePercent
}

func (c *CFishable) GetYield() LootItem {
	return c.yield
}
