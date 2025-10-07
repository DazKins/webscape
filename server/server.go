package server

import (
	"log"
	"net/http"
	"webscape/server/game"
)

func Start() {
	fs := http.FileServer(http.Dir("client"))
	noCacheHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fs.ServeHTTP(w, r)
	})

	game := game.NewGame()

	clientMessageHandler := NewClientMessageHandler(game)

	game.StartUpdateLoop()

	wsServer := NewWsServer()
	wsServer.SetIncomingMessageHandler(clientMessageHandler.HandleMessage)
	wsServer.SetDisconnectHandler(game.HandleLeave)

	game.RegisterBroadcaster(wsServer.Broadcast)
	game.RegisterSender(wsServer.SendToClient)
	http.Handle("/", noCacheHandler)
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
