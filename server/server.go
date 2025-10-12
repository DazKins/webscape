package server

import (
	"io/fs"
	"log"
	"net/http"
	"webscape/server/game"
)

func Start(distFS fs.FS) {
	http.Handle("/", http.FileServer(http.FS(distFS)))

	game := game.NewGame()

	clientMessageHandler := NewClientMessageHandler(game)

	game.StartUpdateLoop()

	wsServer := NewWsServer()
	wsServer.SetIncomingMessageHandler(clientMessageHandler.HandleMessage)
	wsServer.SetDisconnectHandler(game.HandleLeave)

	game.RegisterBroadcaster(wsServer.Broadcast)
	game.RegisterSender(wsServer.SendToClient)
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Println("Starting server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
