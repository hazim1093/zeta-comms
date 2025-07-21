package main

import (
	"os"

	"github.com/rs/zerolog"
)

func main() {
	log := InitLogger("console", "info") // TODO: get from config

	log.Info().Msg("hi")
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
