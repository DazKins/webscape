package server

import (
	"io/fs"
	"log"
	"net/http"
	"webscape/server/command"
	"webscape/server/game"
)

func Start(distFS fs.FS) {
	http.Handle("/", http.FileServer(http.FS(distFS)))

	game := game.NewGame()

	clientCommandHandler := NewClientCommandHandler(game)

	game.StartUpdateLoop()

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
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Println("Starting server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
