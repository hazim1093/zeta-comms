package events

import (
	"context"
	"time"

	"github.com/hazim1093/zeta-comms/internal/clients"
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/rs/zerolog"
)

const (
	SoftwareUpgradeProposalType = "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade"
)

type GovService struct {
	restClient *clients.RESTClient
	config     *config.Config
	log        *zerolog.Logger
}

// ProposalUpdate contains either proposals or an error
type ProposalUpdate struct {
	Proposals []clients.Proposal
	Error     error
}

func NewGovService(cfg *config.Config, logger *zerolog.Logger) *GovService {
	restClient := clients.NewRESTClient(cfg, logger)

	return &GovService{
		restClient: restClient,
		config:     cfg,
		log:        logger,
	}
}

// StartPollingProposals starts polling for software upgrade proposals and returns channels for updates and errors
func (g *GovService) StartPollingProposals(ctx context.Context, network string) chan ProposalUpdate {
	log := g.log.With().Str("network", network).Logger()
	pollInterval := g.config.Networks[network].PollInterval

	log.Info().Msg("Starting to poll software upgrade proposals every " + pollInterval.String())

	// Create a buffered channel to avoid blocking
	updateCh := make(chan ProposalUpdate, 10)

	go g.pollProposals(ctx, network, pollInterval, updateCh, &log)

	log.Info().Msg("Polling started")
	return updateCh
}

func (g *GovService) pollProposals(ctx context.Context, network string, pollInterval time.Duration, updateCh chan ProposalUpdate, log *zerolog.Logger) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	defer close(updateCh)

	// Initial fetch
	proposals, err := g.getSoftwareUpgradeProposals(network)
	if err != nil {
		log.Error().Err(err).Msg("failed to get initial proposals")
		// Send the error to the channel instead of returning
		updateCh <- ProposalUpdate{Error: err}
	} else {
		log.Debug().Msgf("Initial proposals fetched")
		// Send initial proposals to the channel
		updateCh <- ProposalUpdate{Proposals: proposals}
	}

	// Polling loop
	for {
		select {
		case <-ticker.C:
			proposals, err := g.getSoftwareUpgradeProposals(network)
			if err != nil {
				log.Error().Err(err).Msg("failed to get proposals")
				updateCh <- ProposalUpdate{Error: err}
			} else {
				log.Debug().Msgf("Proposals fetched")
				updateCh <- ProposalUpdate{Proposals: proposals}
			}

		case <-ctx.Done():
			log.Info().Msg("Stopping proposal polling due to context cancellation")
			return
		}
	}
}

func (g *GovService) getSoftwareUpgradeProposals(network string) ([]clients.Proposal, error) {
	proposalsResp, err := g.restClient.GetProposals(network)
	if err != nil {
		g.log.Error().Err(err).Msg("failed to get proposals")
		return nil, err
	}

	proposals := filterSoftwareUpgradeProposals(proposalsResp.Proposals)
	return proposals, nil
}

// TODO: look at the logic again
func filterSoftwareUpgradeProposals(proposals []clients.Proposal) []clients.Proposal {
	var filtered []clients.Proposal

	for _, proposal := range proposals {
		for _, message := range proposal.Messages {
			if message.Type == SoftwareUpgradeProposalType {
				filtered = append(filtered, proposal)
				break
			}
		}
	}

	return filtered
}
