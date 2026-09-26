package component

const ComponentIdAttackCooldown = ComponentId("attackcooldown")

// CAttackCooldown belongs to the actor, independently of its current intent or
// weapon. Cancelling or replacing combat must not make the next attack ready.
// It is server-only session state, not an animation phase.
type CAttackCooldown struct {
	nextAttackTick uint64
}

func NewCAttackCooldown(nextAttackTick uint64) *CAttackCooldown {
	return &CAttackCooldown{nextAttackTick: nextAttackTick}
}

func (*CAttackCooldown) GetId() ComponentId { return ComponentIdAttackCooldown }

func (c *CAttackCooldown) GetNextAttackTick() uint64 { return c.nextAttackTick }

func (*CAttackCooldown) transientSave() {}
