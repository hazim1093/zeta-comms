package notifications

import (
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
	"github.com/hazim1093/zeta-comms/pkg/notifiers/discord"
	"github.com/hazim1093/zeta-comms/pkg/notifiers/slack"
	"github.com/hazim1093/zeta-comms/pkg/notifiers/telegram"
	"github.com/rs/zerolog"
)

// Use the Notification type from the models package
type Notification = models.Notification

type NotificationService struct {
	config    *config.Config
	log       *zerolog.Logger
	notifiers map[string]notifiers.Notifier
}

func NewNotificationService(cfg *config.Config, log *zerolog.Logger) *NotificationService {
	service := &NotificationService{
		config:    cfg,
		log:       log,
		notifiers: make(map[string]notifiers.Notifier),
	}

	// Initialize Discord client
	if discordClient, err := discord.InitializeDiscordClient(log, cfg.Notifiers.Discord.BotToken); err == nil {
		service.notifiers["discord"] = discordClient
	} else {
		log.Error().Err(err).Msg("Failed to initialize Discord client")
	}

	// Initialize Telegram client
	if telegramClient, err := telegram.InitializeTelegramClient(log, cfg.Notifiers.Telegram.BotToken); err == nil {
		service.notifiers["telegram"] = telegramClient
	} else {
		log.Error().Err(err).Msg("Failed to initialize Telegram client")
	}

	// Initialize Slack client
	service.notifiers["slack"] = slack.NewSlackClient(log)

	return service
}

func (n *NotificationService) Notify(notification Notification, audience string) {
	log := n.log.With().Str("audience", audience).Logger()

	audienceConfig, ok := n.config.AudienceConfig[audience]
	if !ok {
		log.Error().Msg("No audience config found")

		return
	}

	for platform, channels := range audienceConfig.Channels {
		n.sendToChannels(platform, channels, notification, log)
	}
}

func (n *NotificationService) sendToChannels(platform string, channels []string, notification Notification, log zerolog.Logger) {
	notifier, exists := n.notifiers[platform]
	if !exists {
		log.Error().Msgf("No notifier found for platform: %s", platform)

		return
	}

	for _, channel := range channels {
		err := notifier.Send(channel, notification)
		if err != nil {
			log.Error().
				Err(err).
				Str("platform", platform).
				Str("channel", channel).
				Str("proposal_id", notification.ProposalId).
				Msg("Failed to send notification")

			continue
		}

		log.Info().
			Str("platform", platform).
			Str("proposal_id", notification.ProposalId).
			Msg("Notification sent successfully")
	}
}
