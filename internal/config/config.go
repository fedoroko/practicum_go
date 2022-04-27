package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

type Config interface {
	Flags() *Config
	Env() *Config
}

type ServerConfig struct {
	Address       string        `env:"ADDRESS"`
	Restore       bool          `env:"RESTORE"`
	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Key           string        `env:"KEY"`
	Database      string        `env:"DATABASE_DSN"`
	Debug         bool
}

func (s *ServerConfig) Flags() *ServerConfig {
	flag.StringVar(&s.Address, "a", "127.0.0.1:8080", "Host address")
	flag.BoolVar(&s.Restore, "r", true, "Restore previous db")
	flag.DurationVar(&s.StoreInterval, "i", time.Second*300, "Store interval")
	flag.StringVar(&s.StoreFile, "f", "/tmp/devops-metrics-db.json", "Store file path")
	flag.StringVar(&s.Key, "k", "", "Key for hashing")
	flag.StringVar(&s.Database, "d", "", "Database DSN")
	flag.BoolVar(&s.Debug, "debug", false, "Debug mode")
	flag.Parse()

	return s
}

func (s *ServerConfig) Env() *ServerConfig {
	err := env.Parse(s)
	if err != nil {
		log.Err(err).Send()
	}

	return s
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:       "127.0.0.1:8080",
		Restore:       false,
		StoreInterval: time.Second * 300,
		StoreFile:     "/tmp/devops-metrics-db.json",
	}
}

type AgentConfig struct {
	Address          string        `env:"ADDRESS"`
	PollInterval     time.Duration `env:"POLL_INTERVAL"`
	ReportInterval   time.Duration `env:"REPORT_INTERVAL"`
	ShutdownInterval time.Duration
	ContentType      string
	Key              string `env:"KEY"`
	Debug            bool
}

func (a *AgentConfig) Flags() *AgentConfig {
	flag.StringVar(&a.Address, "a", "127.0.0.1:8080", "Host address")
	flag.DurationVar(&a.PollInterval, "p", time.Second*2, "Poll count interval")
	flag.DurationVar(&a.ReportInterval, "r", time.Second*10, "Report interval")
	flag.StringVar(&a.Key, "k", "", "Key for hashing")
	flag.BoolVar(&a.Debug, "debug", false, "Debug mode - bool")
	flag.Parse()

	return a
}

func (a *AgentConfig) Env() *AgentConfig {
	err := env.Parse(a)
	if err != nil {
		log.Err(err).Send()
	}

	return a
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		Address:          "127.0.0.1:8080",
		PollInterval:     time.Second * 2,
		ReportInterval:   time.Second * 10,
		ShutdownInterval: time.Second * 500,
		ContentType:      "text/plain",
	}
}

type Logger struct {
	*zerolog.Logger
}

func TestLogger() *Logger {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}
	logger := zerolog.New(output).With().Timestamp().Logger()

	return &Logger{
		Logger: &logger,
	}
}

func NewLogger(logger *zerolog.Logger) *Logger {
	return &Logger{
		Logger: logger,
	}
}

func (s *ServerConfig) GetLogger() *Logger {
	logLevel := zerolog.InfoLevel
	if s.Debug {
		logLevel = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}
	logger := zerolog.New(output).With().Timestamp().Logger()

	return NewLogger(&logger)
}

func (a *AgentConfig) GetLogger() *Logger {
	logLevel := zerolog.InfoLevel
	if a.Debug {
		logLevel = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFormatUnix}
	logger := zerolog.New(output).With().Timestamp().Logger()

	return NewLogger(&logger)
}
