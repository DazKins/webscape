package server

import (
	"io/fs"
	"log"
	"net/http"
	"time"
	"webscape/server/command"
	"webscape/server/game"
	"webscape/server/game/world"
)

func Start(distFS fs.FS, gameWorld *world.World, address string, chunkRadius int, tickInterval time.Duration, devMode bool) {
	http.Handle("/", frontendHandler(distFS, devMode))
	if devMode {
		log.Print("Development mode: frontend caching disabled")
	}

	game := game.NewGameWithWorldAndChunkRadius(gameWorld, chunkRadius)

	clientCommandHandler := NewClientCommandHandler(game)
	wsServer := NewWsServer()
	wsServer.SetConnectHandler(game.HandleConnect)
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
	game.StartUpdateLoop(tickInterval)
	http.HandleFunc("/ws", wsServer.HandleWebSocket)

	log.Printf("Starting server on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
}

func frontendHandler(distFS fs.FS, devMode bool) http.Handler {
	files := http.FileServer(http.FS(distFS))
	if !devMode {
		return files
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		files.ServeHTTP(w, r)
	})
}
