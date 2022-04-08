package main

import "github.com/fedoroko/practicum_go/internal/agent"

func main() {
	agent.Run(
		agent.WithPollInterval(2),
		agent.WithReportInterval(10),
		agent.WithShutdownInterval(500),
		agent.WithContentType(agent.ContentTypeJson),
	)
}
