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

func NewGovService(cfg *config.Config, logger *zerolog.Logger) *GovService {
	restClient := clients.NewRESTClient(cfg, logger)

	return &GovService{
		restClient: restClient,
		config:     cfg,
		log:        logger,
	}
}

func (g *GovService) StartPollingProposals(ctx context.Context, network string) {
	log := g.log.With().Str("network", network).Logger()

	pollInterval := g.config.Networks[network].PollInterval

	log.Info().Msg("Starting to poll software upgrade proposals every " + pollInterval.String())

	go g.pollProposals(ctx, network, pollInterval, &log)

	log.Info().Msg("Polling started")
}

func (g *GovService) pollProposals(ctx context.Context, network string, pollInterval time.Duration, log *zerolog.Logger) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Initial fetch
	if proposals, err := g.getSoftwareUpgradeProposals(network); err != nil {
		log.Error().Err(err).Msg("failed to get proposals")
		return
	} else {
		log.Info().Msgf("Initial proposals: %v", proposals)
	}

	// Polling loop
	for {
		select {
		case <-ticker.C:
			if proposals, err := g.getSoftwareUpgradeProposals(network); err != nil {
				log.Error().Err(err).Msg("failed to get proposals")
			} else {
				log.Info().Msgf("Retrieved proposals: %v", proposals)
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
