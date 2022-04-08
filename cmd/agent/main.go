package main

import "github.com/fedoroko/practicum_go/internal/agent"

func main() {
	agent.Run(
		agent.WithPollInterval(1),
		agent.WithReportInterval(4),
		agent.WithShutdownInterval(22),
		agent.WithContentType(agent.ContentTypeJson),
	)
}
