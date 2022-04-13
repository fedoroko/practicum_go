package main

import (
	"log"

	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/server"
)

func main() {
	log.Println("server started")
	defer log.Println("server ended")

	cfg := config.NewServerConfig().Flags().Env()
	server.Run(cfg)
}
