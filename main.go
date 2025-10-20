package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"webscape/server"
)

//go:embed all:client/dist
var DistFS embed.FS

func main() {
	dev := flag.Bool("dev", false, "Use live filesystem instead of embedded files")
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

	server.Start(f)
}
