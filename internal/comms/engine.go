package comms

import (
	"fmt"
	"strconv"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/internal/events"
	"github.com/hazim1093/zeta-comms/internal/storage"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

// mapProposalToNotification converts a zetachain.Proposal to a models.Notification object
func mapProposalToNotification(network string, proposal zetachain.Proposal) models.Notification {
	// Extract upgrade information if available
	var upgradeName, targetHeight string
	for _, msg := range proposal.Messages {
		if msg.Data.Plan.Name != "" {
			upgradeName = msg.Data.Plan.Name
		}
		if msg.Data.Plan.Height != "" {
			targetHeight = msg.Data.Plan.Height
		}
	}

	// Parse vote counts to float64 for calculations
	yesCount, _ := strconv.ParseFloat(proposal.FinalTallyResult.YesCount, 64)
	noCount, _ := strconv.ParseFloat(proposal.FinalTallyResult.NoCount, 64)
	abstainCount, _ := strconv.ParseFloat(proposal.FinalTallyResult.AbstainCount, 64)
	vetoCount, _ := strconv.ParseFloat(proposal.FinalTallyResult.NoWithVetoCount, 64)

	// Calculate total votes
	totalVotes := yesCount + noCount + abstainCount + vetoCount

	// Convert from azeta to ZETA (1 ZETA = 10^18 azeta)
	divisor := 1000000000000000000.0 // 10^18
	yesZeta := yesCount / divisor
	noZeta := noCount / divisor
	abstainZeta := abstainCount / divisor
	vetoZeta := vetoCount / divisor
	totalZeta := totalVotes / divisor

	// Convert to millions for display
	yesMillions := yesZeta / 1000000
	noMillions := noZeta / 1000000
	abstainMillions := abstainZeta / 1000000
	vetoMillions := vetoZeta / 1000000
	totalMillions := totalZeta / 1000000

	// Calculate percentages (avoid division by zero)
	var yesPercentage, noPercentage, abstainPercentage, vetoPercentage float64
	if totalVotes > 0 {
		yesPercentage = (yesCount / totalVotes) * 100
		noPercentage = (noCount / totalVotes) * 100
		abstainPercentage = (abstainCount / totalVotes) * 100
		vetoPercentage = (vetoCount / totalVotes) * 100
	}

	// Format vote strings with both values and percentages
	yesVotesStr := fmt.Sprintf("%.3fM (%.2f%%)", yesMillions, yesPercentage)
	noVotesStr := fmt.Sprintf("%.3fM (%.2f%%)", noMillions, noPercentage)
	abstainVotesStr := fmt.Sprintf("%.3fM (%.2f%%)", abstainMillions, abstainPercentage)
	vetoVotesStr := fmt.Sprintf("%.3fM (%.2f%%)", vetoMillions, vetoPercentage)
	totalVotesStr := fmt.Sprintf("%.3fM", totalMillions)

	// Convert azeta to ZETA in deposit information (1 ZETA = 10^18 azeta)
	// Note: This conversion is only for deposit amounts, not for vote counts
	convertedDeposit := make([]zetachain.Deposit, 0, len(proposal.TotalDeposit))
	for _, deposit := range proposal.TotalDeposit {
		if deposit.Denom == "azeta" {
			// Convert azeta to ZETA (1 ZETA = 10^18 azeta)
			amount, _ := strconv.ParseFloat(deposit.Amount, 64)
			zetaAmount := amount / 1000000000000000000 // 10^18

			// Create new deposit with converted amount
			convertedDeposit = append(convertedDeposit, zetachain.Deposit{
				Denom:  "ZETA",
				Amount: fmt.Sprintf("%.2f", zetaAmount),
			})
		} else {
			// Keep other denominations as is
			convertedDeposit = append(convertedDeposit, deposit)
		}
	}

	// Create and return the notification with enhanced information
	return models.Notification{
		Network:       network,
		ProposalId:    proposal.ProposalId,
		Title:         proposal.Title,
		Summary:       proposal.Summary,
		Status:        proposal.Status,
		UpgradeName:   upgradeName,
		TargetHeight:  targetHeight,
		YesVotes:      yesVotesStr,
		NoVotes:       noVotesStr,
		AbstainVotes:  abstainVotesStr,
		VetoVotes:     vetoVotesStr,
		TotalVotes:    totalVotesStr,
		SubmitTime:    proposal.SubmitTime,
		VotingEndTime: proposal.VotingEndTime,
		Expedited:     proposal.Expedited,
		FailedReason:  proposal.FailedReason,
		TotalDeposit:  convertedDeposit,
	}
}

func (e *CommsEngine) handleProposals(network string, proposals []zetachain.Proposal) {
	e.log.Trace().Msgf("Handling %d proposals for network: %s", len(proposals), network)

	for _, proposal := range proposals {
		isNew := e.isNewProposal(network, proposal.ProposalId)

		if !isNew {
			e.log.Debug().Msgf("Proposal %s is not new", proposal.ProposalId)
			continue
		}

		log.Info().Msgf("Processing new proposal: %s", proposal.ProposalId)

		notification := mapProposalToNotification(network, proposal)

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
