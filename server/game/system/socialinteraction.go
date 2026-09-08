package system

import (
	"webscape/server/game/component"
	"webscape/server/game/model"
)

type socialParticipant struct{ Actor, Target model.EntityId }

// Conversations and shops both keep participants still and facing one another.
func socialParticipants(manager *component.ComponentManager) []socialParticipant {
	result := []socialParticipant{}
	for _, id := range []component.ComponentId{component.ComponentIdActiveConversation, component.ComponentIdTrading} {
		for actor, value := range manager.GetComponent(id) {
			target := value.(interface{ GetTargetEntityId() model.EntityId }).GetTargetEntityId()
			result = append(result, socialParticipant{Actor: actor, Target: target})
		}
	}
	return result
}

func isSocialParticipant(manager *component.ComponentManager, id model.EntityId) bool {
	for _, pair := range socialParticipants(manager) {
		if pair.Actor == id || pair.Target == id {
			return true
		}
	}
	return false
}
