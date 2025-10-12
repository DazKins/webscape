package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"webscape/server"
)

//go:embed all:client/dist
var DistFS embed.FS

func main() {
	flag.Parse()
	var f fs.FS
	var err error
	f, err = fs.Sub(DistFS, "client/dist")
	if err != nil {
		log.Fatal(err)
	}
	server.Start(f)
}
