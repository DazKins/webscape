package game

import (
	"encoding/json"
	"testing"
	"webscape/server/game/component"
	"webscape/server/game/model"
	"webscape/server/message"
)

func TestPlayerLocomotionStaysMovingAcrossStepsAndIdlesAfterFinalStep(t *testing.T) {
	game, _ := setupDistantFishingGame(t)
	playerId := joinPlayer(t, game, "client-1", "One")
	assertLocomotion(t, game, playerId, component.LocomotionPhaseIdle, 0)

	game.HandleMove("client-1", 3, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseIdle, 4)
}

func TestLocomotionReplicationSupportsObserversAndLateInterest(t *testing.T) {
	game, _ := setupDistantFishingGame(t)
	sent := map[string][]message.Message{}
	game.RegisterSender(func(clientId string, msg message.Message) {
		sent[clientId] = append(sent[clientId], msg)
	})
	playerId := joinPlayer(t, game, "client-1", "One")
	joinPlayer(t, game, "client-2", "Two")
	sent = map[string][]message.Message{}

	game.HandleMove("client-1", 3, 1)
	game.update()
	for _, clientId := range []string{"client-1", "client-2"} {
		state, serverTick := locomotionUpdateFor(t, sent[clientId], playerId)
		if state == nil || state.Phase != string(component.LocomotionPhaseMoving) ||
			state.PhaseStartedTick != 1 || serverTick != 1 {
			t.Fatalf("locomotion for %s = %#v at tick %d", clientId, state, serverTick)
		}
	}

	sent["client-1"] = nil
	sent["client-2"] = nil
	game.update()
	for _, clientId := range []string{"client-1", "client-2"} {
		state, _ := locomotionUpdateFor(t, sent[clientId], playerId)
		if state != nil {
			t.Fatalf("continuous movement restarted locomotion for %s: %#v", clientId, state)
		}
	}
	sent["client-3"] = nil
	joinPlayer(t, game, "client-3", "Three")
	state, serverTick := locomotionUpdateFor(t, sent["client-3"], playerId)
	if state == nil || state.Phase != string(component.LocomotionPhaseMoving) ||
		state.PhaseStartedTick != 1 || serverTick != 2 {
		t.Fatalf("late locomotion snapshot = %#v at tick %d", state, serverTick)
	}
}

func TestLocomotionRestartGetsNewPhaseStartedTick(t *testing.T) {
	game, _ := setupDistantFishingGame(t)
	playerId := joinPlayer(t, game, "client-1", "One")
	game.HandleMove("client-1", 1, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseIdle, 2)

	game.HandleMove("client-1", 0, 1)
	game.update()
	assertLocomotion(t, game, playerId, component.LocomotionPhaseMoving, 3)
}

type serializedLocomotionState struct {
	Phase            string `json:"phase"`
	PhaseStartedTick uint64 `json:"phaseStartedTick"`
}

func locomotionUpdateFor(
	t *testing.T,
	messages []message.Message,
	entityId model.EntityId,
) (*serializedLocomotionState, uint64) {
	t.Helper()
	var result *serializedLocomotionState
	var serverTick uint64
	for _, msg := range messages {
		if msg.Metadata.Type != message.MessageTypeGameUpdate {
			continue
		}
		var payload struct {
			Data struct {
				ServerTick uint64 `json:"serverTick"`
				Entities   []struct {
					EntityId    string          `json:"entityId"`
					ComponentId string          `json:"componentId"`
					Data        json.RawMessage `json:"data"`
				} `json:"entities"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(msg.Marshal()), &payload); err != nil {
			t.Fatal(err)
		}
		for _, update := range payload.Data.Entities {
			if update.EntityId != entityId.String() ||
				update.ComponentId != component.ComponentIdLocomotion.String() {
				continue
			}
			state := &serializedLocomotionState{}
			if err := json.Unmarshal(update.Data, state); err != nil {
				t.Fatal(err)
			}
			result = state
			serverTick = payload.Data.ServerTick
		}
	}
	return result, serverTick
}

func assertLocomotion(
	t *testing.T,
	game *Game,
	entityId model.EntityId,
	wantPhase component.LocomotionPhase,
	wantStartedTick uint64,
) {
	t.Helper()
	value := game.componentManager.GetEntityComponent(component.ComponentIdLocomotion, entityId)
	if value == nil {
		t.Fatal("player has no locomotion state")
	}
	locomotion := value.(*component.CLocomotion)
	if locomotion.GetPhase() != wantPhase || locomotion.GetPhaseStartedTick() != wantStartedTick {
		t.Fatalf(
			"locomotion = %q at tick %d, want %q at tick %d",
			locomotion.GetPhase(),
			locomotion.GetPhaseStartedTick(),
			wantPhase,
			wantStartedTick,
		)
	}
}
