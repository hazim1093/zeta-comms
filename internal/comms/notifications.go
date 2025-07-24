package comms

import (
	"fmt"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/discord"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/slack"
	"github.com/hazim1093/zeta-comms/pkg/telegram"
	"github.com/rs/zerolog"
)

// Use the Notification type from the models package
type Notification = models.Notification

type NotificationService struct {
	config   *config.Config
	log      *zerolog.Logger
	Slack    *slack.SlackClient
	Discord  *discord.DiscordClient
	Telegram *telegram.TelegramClient
}

func NewNotificationService(cfg *config.Config, log *zerolog.Logger) *NotificationService {
	// Initialize Discord client
	discordClient, err := discord.NewDiscordClient(log, cfg.AuthConfig.Discord.BotToken)
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

	// Initialize Telegram client
	telegramClient, err := telegram.NewTelegramClient(log, cfg.AuthConfig.Telegram.BotToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Telegram client")
	} else {
		// Connect to Telegram
		err = telegramClient.Connect()
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to Telegram")
		}
	}

	return &NotificationService{
		config:   cfg,
		log:      log,
		Slack:    slack.NewSlackClient(log),
		Discord:  discordClient,
		Telegram: telegramClient,
	}
}

func (n *NotificationService) Notify(notification Notification, audience string) {
	log := n.log.With().Str("audience", audience).Logger()

	log.Info().
		Str("id", notification.ProposalId).
		Str("title", notification.Title).
		Str("status", notification.Status).
		Msg("Processing proposal")

	audienceConfig, ok := n.config.AudienceConfig[audience]
	if !ok {
		log.Error().Msg("No audience config found")
		return
	}

	for _, channelID := range audienceConfig.Channels.Discord {
		n.sendDiscordNotification(channelID, notification)
	}

	for _, webhook := range audienceConfig.Channels.Slack {
		n.sendSlackNotification(webhook, notification)
	}

	for _, chatID := range audienceConfig.Channels.Telegram {
		n.sendTelegramNotification(chatID, notification)
	}
}

func (n *NotificationService) sendSlackNotification(webhook string, notification Notification) {
	n.log.Debug().Msg("Sending Slack notification to webhook")

	// Format the notification into a Slack message
	message := slack.FormatProposalMessage(notification)

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

	n.log.Debug().Msg("Sending Discord notification to channel: " + channelID)

	// Format the notification into a Discord message
	embed := discord.FormatProposalMessage(notification)

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

func (n *NotificationService) sendTelegramNotification(chatID string, notification Notification) {
	if n.Telegram == nil {
		n.log.Warn().Msg("Telegram client not initialized")
		return
	}

	n.log.Debug().Msg("Sending Telegram notification to chat: " + chatID)

	// Format the notification into a Telegram message
	message := telegram.FormatProposalMessage(notification)

	// Send the message to the Telegram chat
	err := n.Telegram.SendMessage(chatID, message, "Markdown")
	if err != nil {
		n.log.Error().
			Err(err).
			Str("chat_id", chatID).
			Str("proposal_id", notification.ProposalId).
			Msg("Failed to send Telegram notification")
		return
	}

	n.log.Info().
		Str("proposal_id", notification.ProposalId).
		Msg("Telegram notification sent successfully")
}
