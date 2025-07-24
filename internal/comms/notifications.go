package comms

import (
	"fmt"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/discord"
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
	config  *config.Config
	log     *zerolog.Logger
	Slack   *slack.SlackClient
	Discord *discord.DiscordClient
}

func NewNotificationService(cfg *config.Config, log *zerolog.Logger) *NotificationService {
	discordClient, err := discord.NewDiscordClient(log, cfg.Discord.BotToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Discord client")
	} else {
		// Connect to Discord
		err = discordClient.Connect()
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to Discord")
		} else {
			discordClient.AddReconnectHandler()
		}
	}

	return &NotificationService{
		config:  cfg,
		log:     log,
		Slack:   slack.NewSlackClient(log),
		Discord: discordClient,
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

	for _, channelID := range audienceConfig.Channels.Discord {
		n.sendDiscordNotification(channelID, notification)
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

func (n *NotificationService) sendDiscordNotification(channelID string, notification Notification) {
	if n.Discord == nil {
		n.log.Warn().Msg("Discord client not initialized")
		return
	}

	n.log.Debug().Msg("Sending Discord notification")

	// Format the notification into a Discord message
	embed := discord.FormatProposalMessage(
		notification.Title,
		notification.Message,
		notification.ProposalId,
		notification.ProposalStatus,
	)

	// Create a simple content message
	content := fmt.Sprintf("New proposal update: %s", notification.Title)

	// Send the message to the Discord channel
	err := n.Discord.SendChannelMessage(channelID, content, embed)
	if err != nil {
		n.log.Error().
			Err(err).
			Str("channel_id", channelID).
			Str("proposal_id", notification.ProposalId).
			Msg("Failed to send Discord notification")
		return
	}

	n.log.Info().
		Str("proposal_id", notification.ProposalId).
		Msg("Discord notification sent successfully")
}
