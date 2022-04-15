package main

import (
	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/server"
)

func main() {
	cfg := config.NewServerConfig().Flags().Env()
	server.Run(cfg)
}
