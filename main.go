package main

import (
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"webscape/server"
	"webscape/server/game/world"
)

//go:embed all:client/dist
var DistFS embed.FS

func main() {
	dev := flag.Bool("dev", false, "Use live filesystem instead of embedded files")
	gameFolder := flag.String("game-folder", "", "Load game content from a folder containing game.json")
	flag.Parse()

	var f fs.FS
	var err error

	if *dev {
		// Development mode: read from actual filesystem
		log.Println("Using live filesystem (development mode)")
		f = os.DirFS("client/dist")
	} else {
		// Production mode: use embedded files
		log.Println("Using embedded filesystem (production mode)")
		f, err = fs.Sub(DistFS, "client/dist")
		if err != nil {
			log.Fatal(err)
		}
	}

	gameWorld, err := loadGameWorld(*gameFolder)
	if err != nil {
		log.Fatal(err)
	}

	server.Start(f, gameWorld)
}

func loadGameWorld(gameFolder string) (*world.World, error) {
	if gameFolder == "" {
		return nil, errors.New("-game-folder is required")
	}

	log.Printf("Using game folder %q", gameFolder)
	return world.LoadFromGameFolder(gameFolder)
}
