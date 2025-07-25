package events

import (
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers/telegram"
	"github.com/rs/zerolog"
)

func StartTelegramBroadcastClient(log *zerolog.Logger, cfg *config.Config) *chan models.BroadcastMessage {
	telegramClient, err := telegram.InitializeTelegramClient(log, cfg.Notifiers.Telegram.BotToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Telegram client")

		return nil
	}

	broadcastChan := make(chan models.BroadcastMessage, 100) // Buffered channel for broadcast messages

	telegramClient.StartPolling(broadcastChan)

	return &broadcastChan
}
