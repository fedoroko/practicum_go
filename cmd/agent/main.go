package main

import "github.com/fedoroko/practicum_go/internal/agent"

func main() {
	agent.Run(
		agent.WithEnv(),
		agent.WithShutdownInterval(500),
		agent.WithContentType(agent.ContentTypeJSON),
	)
}
