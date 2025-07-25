package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
)

// formatNotification creates a formatted Discord embed for a notification
func formatNotification(notification models.Notification) *discordgo.MessageEmbed {
	// Create a rich embed for the notification
	color := getColorForStatus(notification.Status)

	title := "Message from ZetaChain Governance"
	if notification.Title != "" {
		title = fmt.Sprintf("[%s] Proposal #%s: %s", notification.Network, notification.ProposalId, notification.Title)
	}

	embed := &discordgo.MessageEmbed{
		Title:     title,
		Color:     color,
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "ZetaChain Governance",
		},
	}

	// Build description
	description := ""

	if notification.ProposalId != "" {
		description = fmt.Sprintf("**ID:** %s\n**Status:** %s\n\n", notification.ProposalId, notifiers.FormatStatus(notification.Status))
	}

	// Add software upgrade info if available
	if notification.UpgradeName != "" {
		description += fmt.Sprintf("**Upgrade:** %s\n**Target Height:** %s\n\n", notification.UpgradeName, notification.TargetHeight)
	}

	// Add deposit information if available
	if len(notification.TotalDeposit) > 0 {
		description += "**Deposits:**\n"
		for _, deposit := range notification.TotalDeposit {
			description += fmt.Sprintf("• %s %s\n", deposit.Amount, deposit.Denom)
		}

		description += "\n"
	}

	if notification.TotalVotes != "" {
		description += "*Voting Results:*\n"
		description += fmt.Sprintf("• Yes: %s\n", notification.YesVotes)
		description += fmt.Sprintf("• No: %s\n", notification.NoVotes)
		description += fmt.Sprintf("• Abstain: %s\n", notification.AbstainVotes)
		description += fmt.Sprintf("• Veto: %s\n", notification.VetoVotes)
		description += "\n"
		description += "*Total Votes:* " + notification.TotalVotes + "\n"
	}

	description += "\n"

	// Add timeline information
	if !notification.SubmitTime.IsZero() {
		description += fmt.Sprintf("**Submitted:** %s\n", notification.SubmitTime.Format(time.RFC1123))
	}

	if !notification.VotingEndTime.IsZero() {
		description += fmt.Sprintf("**Voting Ends:** %s\n", notification.VotingEndTime.Format(time.RFC1123))
	}

	// Add expedited flag if true
	if notification.Expedited {
		description += "**Expedited:** Yes\n"
	}

	// Add failed reason if available
	if notification.FailedReason != "" {
		description += fmt.Sprintf("**Failed Reason:** %s\n", notification.FailedReason)
	}

	// Add summary with a separator
	description += "\n**Summary:**\n" + notification.Summary

	embed.Description = description

	return embed
}

// getColorForStatus returns a color integer based on proposal status
func getColorForStatus(status string) int {
	switch status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return 0x3AA3E3 // Blue - Action needed
	case "PROPOSAL_STATUS_PASSED":
		return 0x2EB886 // Green - Positive outcome
	case "PROPOSAL_STATUS_REJECTED":
		return 0xE01E5A // Red - Negative outcome
	default:
		return 0x808080 // Gray - Neutral information
	}
}
