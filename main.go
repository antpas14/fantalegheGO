package main

import (
	"fantalegheGO/internal/server"
	"log"
)

func main() {
	myServer := server.NewMyServer()

	if err := myServer.Serve(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
