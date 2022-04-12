package agent

import (
	"flag"
	"time"
)

func parseFlags(cfg *config) {
	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8080", "Host address")
	p := flag.Int("p", 2, "Poll count interval")
	r := flag.Int("r", 10, "Report interval")
	flag.Parse()
	cfg.PollInterval = time.Duration(*p) * time.Second
	cfg.ReportInterval = time.Duration(*r) * time.Second
}
