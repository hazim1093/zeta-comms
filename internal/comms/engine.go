package comms

import (
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
	"github.com/rs/zerolog"
)

type CommsEngine struct {
	config *config.Config
	log    *zerolog.Logger
}

func NewCommsEngine(cfg *config.Config, log *zerolog.Logger) *CommsEngine {
	return &CommsEngine{
		config: cfg,
		log:    log,
	}
}

// ProcessProposalUpdates handles the proposal updates from the channel
func (e *CommsEngine) ProcessProposalUpdates(network string, updateCh <-chan events.ProposalUpdate) {
	log := e.log.With().Str("network", network).Logger()
	log.Trace().Msg("Starting to process proposal updates")

	for update := range updateCh {
		if update.Error != nil {
			log.Error().Err(update.Error).Msg("Error fetching proposals")
			continue
		}

		log.Info().Msgf("Received %d proposals", len(update.Proposals))

		//TODO: check if proposals are new

		// Process each proposal
		e.handleProposals(network, update.Proposals)
	}
	log.Debug().Msg("Proposal update channel closed")
}

func (e *CommsEngine) handleProposals(network string, proposals []zetachain.Proposal) {
	e.log.Trace().Msgf("Handling %d proposals for network: %s", len(proposals), network)

	for _, proposal := range proposals {

		//TODO: check which proposal to send where

		audiences := e.config.Networks[network].Audiences

		for _, audience := range audiences {
			Notify(e.config, e.log, Notification{
				Title:          proposal.Title,
				Message:        proposal.Summary,
				ProposalId:     proposal.ProposalId,
				ProposalStatus: proposal.Status,
			}, audience)
		}
	}
}
