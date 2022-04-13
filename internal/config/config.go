package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config interface {
	Default()
	Flags()
	Env()
}

type ServerConfig struct {
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
}

func (s *ServerConfig) Default() *ServerConfig {
	s.Address = "127.0.0.1:8080"
	s.Restore = false
	s.StoreInterval = time.Second * 300
	s.StoreFile = "/tmp/devops-metrics-db.json"

	return s
}

func (s *ServerConfig) Flags() *ServerConfig {
	flag.StringVar(&s.Address, "a", "127.0.0.1:8080", "Host address")
	flag.BoolVar(&s.Restore, "r", true, "Restore previous db")
	flag.DurationVar(&s.StoreInterval, "i", time.Second*300, "Store interval")
	flag.StringVar(&s.StoreFile, "f", "/tmp/devops-metrics-db.json", "Store file path")
	flag.Parse()

	return s
}

func (s *ServerConfig) Env() *ServerConfig {
	err := env.Parse(s)
	if err != nil {
		log.Println(err)
	}

	return s
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{}
}

type AgentConfig struct {
	Address          string        `env:"ADDRESS"`
	PollInterval     time.Duration `env:"POLL_INTERVAL"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL"`
	ShutdownInterval time.Duration
	ContentType      string
}

func (a *AgentConfig) Default() *AgentConfig {
	a.Address = "127.0.0.1:8080"
	a.PollInterval = time.Second * 2
	a.ReportInterval = time.Second * 10
	a.ShutdownInterval = time.Minute * 10
	a.ContentType = "text/plain"

	return a
}

func (a *AgentConfig) Flags() *AgentConfig {
	flag.StringVar(&a.Address, "a", "127.0.0.1:8080", "Host address")
	flag.DurationVar(&a.PollInterval, "p", time.Second*2, "Poll count interval")
	flag.DurationVar(&a.ReportInterval, "r", time.Second*10, "Report interval")
	flag.Parse()

	return a
}

func (a *AgentConfig) Env() *AgentConfig {
	err := env.Parse(a)
	if err != nil {
		log.Println(err)
	}

	return a
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		ShutdownInterval: time.Second * 500,
		ContentType:      "text/plain",
	}
}
