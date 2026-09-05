package system

import (
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/spatial"
	"webscape/server/game/world"
	"webscape/server/math"
	"webscape/server/util"
)

// Includes planning and the first movement step in a populated 32x32 chunk.
// Each iteration starts a fresh approach so caching cannot hide search cost.
func BenchmarkAttackApproach(b *testing.B) {
	w := world.NewWorld(32, 32)
	manager := component.NewComponentManager()
	index := spatial.NewIndex(w, manager)
	for n := 0; n < 30; n++ {
		manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: n, Y: 15}), component.NewCMetadata(util.JObject{"blocksMovement": util.JBool(n%3 == 0)}))
	}
	target := manager.CreateNewEntity(component.NewCPosition(math.Vec2{X: 24, Y: 24}))
	position := component.NewCPosition(math.Vec2{})
	id := manager.CreateNewEntity(position, component.NewCCombatState(target), component.NewCCombatStats(1, 2, 50, 0, 0, 0, 1, 4, 2))
	s := PathingSystem{SystemBase: SystemBase{ComponentManager: manager}, World: w, SpatialIndex: index}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		position.SetPosition(math.Vec2{})
		manager.SetEntityComponent(id, position)
		manager.SetEntityComponent(id, component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(target)}))
		s.Update()
		if position.GetPosition() == (math.Vec2{}) {
			b.Fatal("approach did not move")
		}
	}
}
