package main

import (
	"github.com/fedoroko/practicum_go/internal/agent"
	"github.com/fedoroko/practicum_go/internal/config"
)

func main() {
	cfg := config.NewAgentConfig().Flags().Env()
	agent.Run(
		cfg,
		agent.WithContentType(agent.ContentTypeJSON),
	)
}
