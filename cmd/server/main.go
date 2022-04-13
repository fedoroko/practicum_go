package main

import (
	"log"

	"github.com/fedoroko/practicum_go/internal/server"
)

func main() {
	log.Println("server started")
	defer log.Println("server ended")
	server.Run(
		server.WithEnv(),
	)
}
