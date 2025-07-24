package comms

import (
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/slack"
	"github.com/rs/zerolog"
)

type Notification struct {
	Title          string `json:"title"`
	Message        string `json:"message"`
	ProposalId     string `json:"proposal_id"`
	ProposalStatus string `json:"proposal_status"`
}

type NotificationService struct {
	config *config.Config
	log    *zerolog.Logger
	Slack  *slack.SlackClient
}

func NewNotificationService(cfg *config.Config, log *zerolog.Logger) *NotificationService {
	return &NotificationService{
		config: cfg,
		log:    log,
		Slack:  slack.NewSlackClient(cfg, log),
	}
}

func (n *NotificationService) Notify(notification Notification, audience string) {
	log := n.log.With().Str("audience", audience).Logger()
	log.Trace().Msg("Sending notification")

	log.Info().
		Str("id", notification.ProposalId).
		Str("title", notification.Title).
		Str("status", notification.ProposalStatus).
		Msg("Processing proposal")

	audienceConfig, ok := n.config.AudienceConfig[audience]
	if !ok {
		log.Warn().Str("audience", audience).Msg("No audience config found")
		return
	}

	for _, channel := range audienceConfig.Channels.Discord {
		log.Info().Msg("Sending Discord notification: " + channel)
		//go sendDiscordNotification(channel, notification)
	}

	for _, webhook := range audienceConfig.Channels.Slack {
		n.sendSlackNotification(webhook, notification)
	}

	for _, channel := range audienceConfig.Channels.Telegram {
		log.Info().Msg("Sending Telegram notification: " + channel)
		//go sendTelegramNotification(channel, notification)
	}
}

func (n *NotificationService) sendSlackNotification(webhook string, notification Notification) {
	n.log.Debug().Msg("Sending Slack notification")

	// Format the notification into a Slack message
	message := slack.FormatProposalMessage(
		notification.Title,
		notification.Message,
		notification.ProposalId,
		notification.ProposalStatus,
	)

	// Send the message to the Slack webhook
	err := n.Slack.SendWebhookMessage(webhook, message)
	if err != nil {
		n.log.Error().
			Err(err).
			Str("webhook", webhook).
			Str("proposal_id", notification.ProposalId).
			Msg("Failed to send Slack notification")
		return
	}

	n.log.Info().
		Str("proposal_id", notification.ProposalId).
		Msg("Slack notification sent successfully")
}
