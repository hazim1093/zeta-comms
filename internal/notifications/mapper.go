package notifications

import (
	"fmt"
	"strconv"

	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
)

func MapFromProposal(network string, proposal zetachain.Proposal) models.Notification {
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
