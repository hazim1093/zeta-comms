package telegram

import (
	"fmt"
	"time"

	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
)

// formatNotification creates a formatted Telegram message for a notification
func formatNotification(notification models.Notification) string {
	formattedMessage := fmt.Sprintf("*[%s]* *Proposal* %s: %s\n\n", notification.Network, notification.ProposalId, notification.Title)
	formattedMessage += fmt.Sprintf("*ID:* %s\n", notification.ProposalId)
	formattedMessage += fmt.Sprintf("*Status:* %s\n\n", notifiers.FormatStatus(notification.Status))

	// Add software upgrade info if available
	if notification.UpgradeName != "" {
		formattedMessage += fmt.Sprintf("*Upgrade:* %s\n*Target Height:* %s\n\n", notification.UpgradeName, notification.TargetHeight)
	}

	// Add deposit information if available
	if len(notification.TotalDeposit) > 0 {
		formattedMessage += "*Deposits:*\n"
		for _, deposit := range notification.TotalDeposit {
			formattedMessage += fmt.Sprintf("• %s %s\n", deposit.Amount, deposit.Denom)
		}

		formattedMessage += "\n"
	}

	formattedMessage += "*Voting Results:*\n"
	if notification.TotalVotes != "" {
		formattedMessage += fmt.Sprintf("• Yes: %s\n", notification.YesVotes)
		formattedMessage += fmt.Sprintf("• No: %s\n", notification.NoVotes)
		formattedMessage += fmt.Sprintf("• Abstain: %s\n", notification.AbstainVotes)
		formattedMessage += fmt.Sprintf("• Veto: %s\n", notification.VetoVotes)
		formattedMessage += "\n"
		formattedMessage += "*Total Votes:* " + notification.TotalVotes + "\n"
	} else {
		formattedMessage += "No voting results available.\n"
	}

	formattedMessage += "\n"

	// Add timeline information
	if !notification.SubmitTime.IsZero() {
		formattedMessage += fmt.Sprintf("*Submitted:* %s\n", notification.SubmitTime.Format(time.RFC1123))
	}

	if !notification.VotingEndTime.IsZero() {
		formattedMessage += fmt.Sprintf("*Voting Ends:* %s\n", notification.VotingEndTime.Format(time.RFC1123))
	}

	// Add expedited flag if true
	if notification.Expedited {
		formattedMessage += "*Expedited:* Yes\n"
	}

	// Add failed reason if available
	if notification.FailedReason != "" {
		formattedMessage += fmt.Sprintf("*Failed Reason:* %s\n", notification.FailedReason)
	}

	// Add summary with a separator
	formattedMessage += "\n*Summary:*\n" + notification.Summary
	formattedMessage += fmt.Sprintf("\n\n_Updated at: %s_", time.Now().Format(time.RFC1123))

	return formattedMessage
}
