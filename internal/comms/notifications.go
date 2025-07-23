package comms

import (
	"fmt"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/rs/zerolog"
)

type Notification struct {
	Title          string `json:"title"`
	Message        string `json:"message"`
	ProposalId     string `json:"proposal_id"`
	ProposalStatus string `json:"proposal_status"`
}

func Notify(config *config.Config, logger *zerolog.Logger, notification Notification, audience string) {
	log := logger.With().Str("audience", audience).Logger()
	log.Trace().Msg("Sending notification")

	log.Info().
		Str("id", notification.ProposalId).
		Str("title", notification.Title).
		Str("status", notification.ProposalStatus).
		Msg("Processing proposal")

	switch notification.ProposalStatus {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		// Alert that voting is open
		fmt.Printf("ALERT: Proposal %s is open for voting\n", notification.ProposalId)
	case "PROPOSAL_STATUS_PASSED":
		// Prepare for the upgrade
		fmt.Printf("ALERT: Proposal %s has passed - prepare for upgrade\n", notification.ProposalId)
	case "PROPOSAL_STATUS_REJECTED":
		// Log the rejection
		fmt.Printf("INFO: Proposal %s was rejected\n", notification.ProposalId)
	}

	audienceConfig, ok := config.AudienceConfig[audience]
	if !ok {
		log.Warn().Str("audience", audience).Msg("No audience config found")
		return
	}

	for _, channel := range audienceConfig.Channels.Discord {
		log.Info().Msg("Sending Discord notification: " + channel)
		//go sendDiscordNotification(channel, notification)
	}
	for _, webhook := range audienceConfig.Channels.Slack {
		log.Info().Msg("Sending Slack notification: " + webhook)
		//go sendSlackNotification(webhook, notification)
	}
	for _, channel := range audienceConfig.Channels.Telegram {
		log.Info().Msg("Sending Telegram notification: " + channel)
		//go sendTelegramNotification(channel, notification)
	}
}
