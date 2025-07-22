package main

import (
	"os"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/rs/zerolog"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		panic(err)
	}

	log := InitLogger(cfg.Logging.Format, cfg.Logging.Level)

	log.Debug().Interface("config", cfg).Msg("config loaded")

	govService := events.NewGovService(cfg, &log)
	govService.StartPollingProposals("testnet")

	// Keep the main function running to allow the polling to continue
	select {}
}

func InitLogger(logFormat string, globalLevel string) zerolog.Logger {
	logLevel, err := zerolog.ParseLevel(globalLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	var logger zerolog.Logger

	switch logFormat {
	case "console", "text":
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	case "json":
		fallthrough
	default:
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return logger
}
