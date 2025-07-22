package events

import (
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

func (g *GovService) GetSoftwareUpgradeProposals(network string) ([]clients.Proposal, error) {
	proposalsResp, err := g.restClient.GetProposals(network)
	if err != nil {
		g.log.Error().Err(err).Msg("failed to get proposals")
		return nil, err
	}

	proposals := filterSoftwareUpgradeProposals(proposalsResp.Proposals)
	return proposals, nil
}

func (g *GovService) StartPollingProposals(network string) {
	log := g.log.With().Str("network", network).Logger()

	pollInterval := g.config.Networks[network].PollInterval

	duration, err := time.ParseDuration(pollInterval)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse poll interval, failed to start polling")
		return
	}

	log.Info().Msg("Starting to poll software upgrade proposals every " + duration.String())

	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		if proposals, err := g.GetSoftwareUpgradeProposals(network); err != nil {
			log.Error().Err(err).Msg("failed to get proposals")
			return
		} else {
			log.Info().Msgf("Initial proposals: %v", proposals)
		}

		for range ticker.C {
			if proposals, err := g.GetSoftwareUpgradeProposals(network); err != nil {
				log.Error().Err(err).Msg("failed to get proposals")
			} else {
				log.Info().Msgf("Retrieved proposals: %v", proposals)
			}
		}

	}()

	log.Info().Msg("Polling started")
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
