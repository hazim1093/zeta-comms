package comms

import (
	"strconv"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/hazim1093/zeta-comms/internal/storage"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
	"github.com/rs/zerolog"
)

type CommsEngine struct {
	config              *config.Config
	log                 *zerolog.Logger
	notificationService *NotificationService
	storageService      *storage.StorageService
}

func NewCommsEngine(cfg *config.Config, log *zerolog.Logger) *CommsEngine {
	return &CommsEngine{
		config:              cfg,
		log:                 log,
		notificationService: NewNotificationService(cfg, log),
		storageService:      storage.NewStorageService(cfg, log),
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

		e.handleProposals(network, update.Proposals)
	}
	log.Debug().Msg("Proposal update channel closed")
}

func (e *CommsEngine) handleProposals(network string, proposals []zetachain.Proposal) {
	e.log.Trace().Msgf("Handling %d proposals for network: %s", len(proposals), network)

	for _, proposal := range proposals {
		isNew := e.isNewProposal(network, proposal.ProposalId)

		if !isNew {
			e.log.Debug().Msgf("Proposal %s is not new", proposal.ProposalId)
			return
		}

		audiences := e.config.Networks[network].Audiences

		for _, audience := range audiences {
			notification := Notification{
				Title:          proposal.Title,
				Message:        proposal.Summary,
				ProposalId:     proposal.ProposalId,
				ProposalStatus: proposal.Status,
			}

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
