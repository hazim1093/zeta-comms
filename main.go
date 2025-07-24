package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hazim1093/zeta-comms/internal/comms"
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

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	//----------------------------------------
	startZetaComms(ctx, cfg, &log)
	//----------------------------------------

	// Wait for termination signal
	sig := <-sigCh
	log.Info().Msgf("Received signal %v, shutting down...", sig)
	cancel() // This will propagate cancellation to the polling goroutine

	log.Info().Msg("Shutdown complete")
}

func startZetaComms(ctx context.Context, cfg *config.Config, log *zerolog.Logger) {
	govService := events.NewGovService(cfg, log)
	commsEngine := comms.NewCommsEngine(cfg, log)

	networks := cfg.Networks
	for network := range networks {
		proposalsChannel := govService.StartPollingProposals(ctx, network)
		if proposalsChannel == nil {
			log.Error().Msgf("Failed to start polling for network: %s", network)

			continue
		}

		go commsEngine.ProcessProposalUpdates(network, proposalsChannel)
	}
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
