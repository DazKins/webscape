package component

const ComponentIdRewardDrop = ComponentId("rewarddrop")

type CRewardDrop struct{}

func NewCRewardDrop() *CRewardDrop {
	return &CRewardDrop{}
}

func (c *CRewardDrop) GetId() ComponentId {
	return ComponentIdRewardDrop
}
