package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
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

	// Allow some time for cleanup
	log.Info().Msg("Shutdown complete")
}

func startZetaComms(ctx context.Context, cfg *config.Config, log *zerolog.Logger) {
	// Start the governance service and get the proposal update channel
	govService := events.NewGovService(cfg, log)
	testnetUpdateChannel := govService.StartPollingProposals(ctx, "testnet")

	// Process proposal updates in a separate goroutine
	go processProposalUpdates(testnetUpdateChannel, log)
}

// processProposalUpdates handles the proposal updates from the channel
func processProposalUpdates(updateCh <-chan events.ProposalUpdate, log *zerolog.Logger) {
	for update := range updateCh {
		if update.Error != nil {
			log.Error().Err(update.Error).Msg("Error fetching proposals")
			continue
		}

		log.Info().Msgf("Received %d proposals", len(update.Proposals))

		// Process each proposal
		for _, proposal := range update.Proposals {
			handleProposal(proposal, log)
		}
	}
	log.Debug().Msg("Proposal update channel closed")
}

// handleProposal processes a single proposal
func handleProposal(proposal zetachain.Proposal, log *zerolog.Logger) {
	log.Info().
		Str("id", proposal.ProposalId).
		Str("title", proposal.Title).
		Str("status", proposal.Status).
		Msg("Processing proposal")

	switch proposal.Status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		// Alert that voting is open
		fmt.Printf("ALERT: Proposal %s is open for voting\n", proposal.ProposalId)
	case "PROPOSAL_STATUS_PASSED":
		// Prepare for the upgrade
		fmt.Printf("ALERT: Proposal %s has passed - prepare for upgrade\n", proposal.ProposalId)
	case "PROPOSAL_STATUS_REJECTED":
		// Log the rejection
		fmt.Printf("INFO: Proposal %s was rejected\n", proposal.ProposalId)
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
