package server

import (
	"io/fs"
	"log"
	"net/http"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/world"
)

func Start(distFS fs.FS, gameWorld *world.World, address string, chunkRadius int) {
	http.Handle("/", http.FileServer(http.FS(distFS)))

	game := game.NewGameWithWorldAndChunkRadius(gameWorld, chunkRadius)

	clientCommandHandler := NewClientCommandHandler(game)
	wsServer := NewWsServer()
	wsServer.SetIncomingMessageHandler(func(clientID string, message string) {
		command, err := command.Unmarshal(message)
		if err != nil {
			log.Printf("error unmarshalling command: %v", err)
			return
		}
		clientCommandHandler.HandleCommand(clientID, command)
	})
	wsServer.SetDisconnectHandler(game.HandleLeave)

	game.RegisterBroadcaster(wsServer.Broadcast)
	game.RegisterSender(wsServer.SendToClient)
	game.StartUpdateLoop()
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Printf("Starting server on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}
