package game

import (
	"sync"
	"testing"
	"webscape/server/game/model"
	"webscape/server/game/world"
	"webscape/server/message"
)

func TestCommandsCanRunConcurrentlyWithUpdates(t *testing.T) {
	gameWorld, err := world.LoadFromGameFS(interestWorldFS())
	if err != nil {
		t.Fatal(err)
	}

	game := NewGameWithWorldAndChunkRadius(gameWorld, 0)
	game.RegisterSender(func(string, message.Message) {})
	game.HandleJoin("client", model.NewEntityId(), "player")

	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for range 100 {
			game.HandleMove("client", 1, 1)
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for range 100 {
			game.update()
		}
	}()

	close(start)
	workers.Wait()
}
