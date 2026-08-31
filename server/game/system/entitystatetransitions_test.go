package system

import (
	"strings"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/math"
	"webscape/server/util"
)

type fixedTickSource uint64

func (s fixedTickSource) CurrentTick() uint64 {
	return uint64(s)
}

func TestEntityStateTransitionsInstallExclusiveStates(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*EntityStateTransitions, model.EntityId, model.EntityId)
		want       component.ComponentId
	}{
		{
			name: "fishing",
			transition: func(transitions *EntityStateTransitions, entityId, targetId model.EntityId) {
				if !transitions.BeginFishing(entityId, component.NewCFishing(targetId, 7, math.Vec2{})) {
					t.Fatal("fishing transition was rejected")
				}
			},
			want: component.ComponentIdFishing,
		},
		{
			name: "woodcutting",
			transition: func(transitions *EntityStateTransitions, entityId, targetId model.EntityId) {
				if !transitions.BeginWoodcutting(entityId, component.NewCWoodcutting(targetId, 7, math.Vec2{})) {
					t.Fatal("woodcutting transition was rejected")
				}
			},
			want: component.ComponentIdWoodcutting,
		},
		{
			name: "combat",
			transition: func(transitions *EntityStateTransitions, entityId, targetId model.EntityId) {
				transitions.BeginCombat(entityId, component.NewCCombatState(targetId))
			},
			want: component.ComponentIdCombatState,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := component.NewComponentManager()
			targetId := model.NewEntityId()
			entityId := manager.CreateNewEntity(
				component.NewCLocomotion(component.LocomotionPhaseIdle, 1),
				component.NewCHealth(10, 10),
				component.NewCActiveConversation("conversation", targetId, "start"),
				component.NewCCombatState(targetId),
				component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
				component.NewCInteracting(targetId, component.InteractionOptionTalk),
				component.NewCWoodcutting(targetId, 7, math.Vec2{}),
				component.NewCFishing(targetId, 1, math.Vec2{}),
			)
			transitions := NewEntityStateTransitions(manager, fixedTickSource(7))

			test.transition(transitions, entityId, targetId)

			for _, componentId := range []component.ComponentId{
				component.ComponentIdActiveConversation,
				component.ComponentIdCombatState,
				component.ComponentIdPathing,
				component.ComponentIdInteracting,
				component.ComponentIdWoodcutting,
				component.ComponentIdFishing,
			} {
				present := manager.GetEntityComponent(componentId, entityId) != nil
				if present != (componentId == test.want) {
					t.Fatalf("component %q present = %v, want %v", componentId, present, componentId == test.want)
				}
			}
			locomotion := manager.GetEntityComponent(component.ComponentIdLocomotion, entityId).(*component.CLocomotion)
			if locomotion.GetPhase() != component.LocomotionPhaseIdle {
				t.Fatalf("locomotion = %q, want idle", locomotion.GetPhase())
			}
			if err := ValidateEntityState(manager, entityId); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEntityStateTransitionsReplaceConflictingPlayerIntent(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*EntityStateTransitions, model.EntityId, model.EntityId)
		want       []component.ComponentId
	}{
		{
			name: "manual pathing",
			transition: func(transitions *EntityStateTransitions, entityId, _ model.EntityId) {
				transitions.BeginPathing(entityId, component.NewCPathing(component.PathingTarget{
					Position: util.OptionalSome(math.Vec2{X: 2}),
				}))
			},
			want: []component.ComponentId{component.ComponentIdPathing},
		},
		{
			name: "interaction",
			transition: func(transitions *EntityStateTransitions, entityId, targetId model.EntityId) {
				transitions.BeginInteraction(
					entityId,
					component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
					component.NewCInteracting(targetId, component.InteractionOptionTalk),
				)
			},
			want: []component.ComponentId{
				component.ComponentIdPathing,
				component.ComponentIdInteracting,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := component.NewComponentManager()
			targetId := model.NewEntityId()
			entityId := manager.CreateNewEntity(
				component.NewCLocomotion(component.LocomotionPhaseIdle, 1),
				component.NewCActiveConversation("conversation", targetId, "start"),
				component.NewCCombatState(targetId),
				component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
				component.NewCInteracting(targetId, component.InteractionOptionTalk),
				component.NewCWoodcutting(targetId, 7, math.Vec2{}),
				component.NewCFishing(targetId, 1, math.Vec2{}),
			)
			transitions := NewEntityStateTransitions(manager, fixedTickSource(7))

			test.transition(transitions, entityId, targetId)

			want := make(map[component.ComponentId]bool)
			for _, componentId := range test.want {
				want[componentId] = true
			}
			for _, componentId := range []component.ComponentId{
				component.ComponentIdActiveConversation,
				component.ComponentIdCombatState,
				component.ComponentIdPathing,
				component.ComponentIdInteracting,
				component.ComponentIdWoodcutting,
				component.ComponentIdFishing,
			} {
				present := manager.GetEntityComponent(componentId, entityId) != nil
				if present != want[componentId] {
					t.Fatalf("component %q present = %v, want %v", componentId, present, want[componentId])
				}
			}
			if err := ValidateEntityState(manager, entityId); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEntityStateTransitionsRejectTemporalActionsOnMovementTick(t *testing.T) {
	manager := component.NewComponentManager()
	targetId := model.NewEntityId()
	locomotion := component.NewCLocomotion(component.LocomotionPhaseIdle, 1)
	locomotion.MarkMoving(7)
	entityId := manager.CreateNewEntity(locomotion)
	transitions := NewEntityStateTransitions(manager, fixedTickSource(7))

	if transitions.BeginFishing(entityId, component.NewCFishing(targetId, 7, math.Vec2{})) {
		t.Fatal("fishing transition accepted movement from the current tick")
	}
	if transitions.BeginWoodcutting(entityId, component.NewCWoodcutting(targetId, 7, math.Vec2{})) {
		t.Fatal("woodcutting transition accepted movement from the current tick")
	}
	if transitions.SettleForInteraction(entityId) {
		t.Fatal("interaction settled on the movement tick")
	}
}

func TestEntityStateTransitionsClearStateForDeath(t *testing.T) {
	manager := component.NewComponentManager()
	targetId := model.NewEntityId()
	locomotion := component.NewCLocomotion(component.LocomotionPhaseIdle, 1)
	locomotion.MarkMoving(7)
	entityId := manager.CreateNewEntity(
		locomotion,
		component.NewCActiveConversation("conversation", targetId, "start"),
		component.NewCCombatState(targetId),
		component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
		component.NewCInteracting(targetId, component.InteractionOptionTalk),
		component.NewCWoodcutting(targetId, 7, math.Vec2{}),
		component.NewCFishing(targetId, 1, math.Vec2{}),
	)
	transitions := NewEntityStateTransitions(manager, fixedTickSource(8))

	transitions.HandleDeath(entityId)

	for _, componentId := range incompatibleComponents[transitionDeath] {
		if manager.GetEntityComponent(componentId, entityId) != nil {
			t.Fatalf("death retained component %q", componentId)
		}
	}
	locomotionState := manager.GetEntityComponent(component.ComponentIdLocomotion, entityId).(*component.CLocomotion)
	if locomotionState.GetPhase() != component.LocomotionPhaseIdle ||
		locomotionState.GetPhaseStartedTick() != 8 {
		t.Fatalf("death locomotion = %q at %d", locomotionState.GetPhase(), locomotionState.GetPhaseStartedTick())
	}
}

func TestValidateEntityStateRejectsImpossibleCombinations(t *testing.T) {
	tests := []struct {
		name       string
		components []component.Component
		want       string
	}{
		{
			name: "moving and fishing",
			components: []component.Component{
				movingLocomotion(4),
				component.NewCFishing(model.NewEntityId(), 4, math.Vec2{}),
			},
			want: "fishing without idle locomotion",
		},
		{
			name: "fishing and combat",
			components: []component.Component{
				component.NewCLocomotion(component.LocomotionPhaseIdle, 4),
				component.NewCFishing(model.NewEntityId(), 4, math.Vec2{}),
				component.NewCCombatState(model.NewEntityId()),
			},
			want: "fishing with incompatible combatstate state",
		},
		{
			name: "woodcutting and pathing",
			components: []component.Component{
				component.NewCLocomotion(component.LocomotionPhaseIdle, 4),
				component.NewCWoodcutting(model.NewEntityId(), 4, math.Vec2{}),
				component.NewCPathing(component.PathingTarget{Position: util.OptionalSome(math.Vec2{})}),
			},
			want: "woodcutting with incompatible pathing state",
		},
		{
			name: "combat and interaction",
			components: []component.Component{
				component.NewCCombatState(model.NewEntityId()),
				component.NewCInteracting(model.NewEntityId(), component.InteractionOptionTalk),
			},
			want: "combat with incompatible interacting state",
		},
		{
			name: "conversation and pathing",
			components: []component.Component{
				component.NewCActiveConversation("conversation", model.NewEntityId(), "start"),
				component.NewCPathing(component.PathingTarget{Position: util.OptionalSome(math.Vec2{})}),
			},
			want: "conversing with incompatible pathing state",
		},
		{
			name: "dead and fishing",
			components: []component.Component{
				component.NewCHealth(10, 0),
				component.NewCLocomotion(component.LocomotionPhaseIdle, 4),
				component.NewCFishing(model.NewEntityId(), 4, math.Vec2{}),
			},
			want: "dead entity",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := component.NewComponentManager()
			entityId := manager.CreateNewEntity(test.components...)
			err := ValidateEntityState(manager, entityId)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateEntityState() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestValidateEntityStateAllowsCompatibleCombinations(t *testing.T) {
	targetId := model.NewEntityId()
	tests := []struct {
		name       string
		components []component.Component
	}{
		{
			name: "moving toward interaction",
			components: []component.Component{
				movingLocomotion(4),
				component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
				component.NewCInteracting(targetId, component.InteractionOptionTalk),
			},
		},
		{
			name: "combat pursuit",
			components: []component.Component{
				movingLocomotion(4),
				component.NewCPathing(component.PathingTarget{EntityId: util.OptionalSome(targetId)}),
				component.NewCCombatState(targetId),
			},
		},
		{
			name: "idle fishing",
			components: []component.Component{
				component.NewCLocomotion(component.LocomotionPhaseIdle, 4),
				component.NewCFishing(targetId, 4, math.Vec2{}),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := component.NewComponentManager()
			entityId := manager.CreateNewEntity(test.components...)
			if err := ValidateEntityState(manager, entityId); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func movingLocomotion(tick uint64) *component.CLocomotion {
	locomotion := component.NewCLocomotion(component.LocomotionPhaseIdle, 0)
	locomotion.MarkMoving(tick)
	return locomotion
}
