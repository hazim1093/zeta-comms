package comms

import (
	"strconv"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/hazim1093/zeta-comms/internal/notifications"
	"github.com/hazim1093/zeta-comms/internal/storage"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
	"github.com/rs/zerolog"
)

type CommsEngine struct {
	config              *config.Config
	log                 *zerolog.Logger
	notificationService *notifications.NotificationService
	storageService      *storage.StorageService
}

func NewCommsEngine(cfg *config.Config, log *zerolog.Logger) *CommsEngine {
	return &CommsEngine{
		config:              cfg,
		log:                 log,
		notificationService: notifications.NewNotificationService(cfg, log),
		storageService:      storage.NewStorageService(cfg, log),
	}
}

// ProcessProposalUpdates handles the proposal updates from the channel
func (e *CommsEngine) ProcessProposalUpdates(network string, updateCh <-chan events.ProposalUpdate) {
	log := e.log.With().Str("network", network).Logger()

	for update := range updateCh {
		if update.Error != nil {
			log.Error().Err(update.Error).Msg("Error fetching proposals")

			continue
		}

		e.handleProposals(network, update.Proposals)
	}

	log.Debug().Msg("Proposal update channel closed")
}

func (e *CommsEngine) handleProposals(network string, proposals []zetachain.Proposal) {
	log := e.log.With().Str("network", network).Logger()
	log.Trace().Msgf("Handling %d proposals for network: %s", len(proposals), network)

	for _, proposal := range proposals {
		isNew := e.isNewProposal(network, proposal.ProposalId)
		if !isNew {
			log.Debug().Msgf("Proposal %s is not new", proposal.ProposalId)

			continue
		}

		log.Info().Str("proposal_id", proposal.ProposalId).Msg("Processing new proposal")

		notification := notifications.MapFromProposal(network, proposal)

		audiences := e.config.Networks[network].Audiences
		for _, audience := range audiences {
			e.notificationService.Notify(notification, audience)
		}

		e.storeLastProcessedProposalID(network, proposal.ProposalId)
	}
}

func (e *CommsEngine) isNewProposal(network string, proposalId string) bool {
	lastProcessedID, err := e.storageService.GetLastProcessedProposalID(network)
	if err != nil {
		e.log.Error().Err(err).Msg("Error getting last processed proposal ID")

		return false
	}

	// If no last processed ID, this is a new proposal
	if lastProcessedID == "" {
		return true
	}

	// Convert strings to integers for comparison
	lastProcessedInt, err1 := strconv.ParseInt(lastProcessedID, 10, 64)
	proposalInt, err2 := strconv.ParseInt(proposalId, 10, 64)

	if err1 != nil || err2 != nil {
		// If conversion fails, fall back to string comparison
		e.log.Warn().Msg("Failed to convert proposal IDs to integers, falling back to string comparison")

		return proposalId != lastProcessedID
	}

	// Check if proposal ID is greater than last processed
	return proposalInt > lastProcessedInt
}

func (e *CommsEngine) storeLastProcessedProposalID(network string, proposalId string) {
	err := e.storageService.StoreLastProcessedProposalID(network, proposalId)
	if err != nil {
		e.log.Error().Err(err).Msg("Error storing last processed proposal ID")

		return
	}
}
