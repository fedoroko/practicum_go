package main

import (
	"github.com/fedoroko/practicum_go/internal/config"
	"github.com/fedoroko/practicum_go/internal/server"
)

func main() {
	cfg := config.NewServerConfig().Flags().Env()
	logger := cfg.GetLogger()

	logger.Debug().Interface("Config", cfg).Send()
	logger.Info().Msg("Server start")
	defer logger.Info().Msg("Server closed")

	server.Run(cfg, logger)
}
