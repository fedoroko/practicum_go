package main

import (
	"github.com/fedoroko/practicum_go/internal/agent"
	"github.com/fedoroko/practicum_go/internal/config"
)

func main() {
	cfg := config.NewAgentConfig().Flags().Env()
	logger := cfg.GetLogger()

	logger.Debug().Interface("Config", cfg).Send()
	logger.Info().Msg("Agent start")
	defer logger.Info().Msg("Agent closed")
	agent.Run(
		cfg,
		logger,
		agent.WithContentType(agent.ContentTypeJSON),
	)
}
