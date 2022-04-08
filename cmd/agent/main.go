package main

import "github.com/fedoroko/practicum_go/internal/agent"

func main() {
	agent.Run(
		agent.WithPollInterval(1),
		agent.WithReportInterval(2),
		agent.WithShutdownInterval(500),
		agent.WithContentType(agent.ContentTypeJSON),
	)
}
