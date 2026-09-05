package main

import (
	"log"
	"os"
	"time"
	"webscape/server"
	"webscape/server/config"
	"webscape/server/game/world"
)

const configPath = "config.json"

func main() {
	runtimeConfig, err := config.LoadFromFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using startup config %q", configPath)

	gameWorld, err := world.LoadFromGameFolder(runtimeConfig.Game.Folder)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using game folder %q", runtimeConfig.Game.Folder)
	clientFolderInfo, err := os.Stat(runtimeConfig.Client.Folder)
	if err != nil {
		log.Fatalf("read client folder %q: %v", runtimeConfig.Client.Folder, err)
	}
	if !clientFolderInfo.IsDir() {
		log.Fatalf("client folder %q is not a directory", runtimeConfig.Client.Folder)
	}
	log.Printf("Using client folder %q", runtimeConfig.Client.Folder)

	server.Start(
		os.DirFS(runtimeConfig.Client.Folder),
		gameWorld,
		runtimeConfig.Server.Address,
		runtimeConfig.Streaming.ChunkRadius,
		time.Duration(runtimeConfig.Server.TickIntervalMs)*time.Millisecond,
	)
}
