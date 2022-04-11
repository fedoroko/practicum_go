package main

import "github.com/fedoroko/practicum_go/internal/agent"

func main() {
	agent.Run(
		agent.WithEnv(),
		agent.WithContentType(agent.ContentTypeJSON),
	)
}
