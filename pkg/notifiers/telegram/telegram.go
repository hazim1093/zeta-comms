package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
	"github.com/rs/zerolog"
)

// TelegramClient handles communication with Telegram API
type TelegramClient struct {
	log *zerolog.Logger
	bot *tgbotapi.BotAPI
}

// Ensure TelegramClient implements the notifiers.Notifier interface
var _ notifiers.Notifier = (*TelegramClient)(nil)

func InitializeTelegramClient(logger *zerolog.Logger, botToken string) (*TelegramClient, error) {
	log := logger.With().Str("service", "telegramClient").Logger()

	// Create a new Telegram bot API client
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Telegram bot: %w", err)
	}

	client := &TelegramClient{
		log: &log,
		bot: bot,
	}

	if err = client.Connect(); err != nil {
		return nil, err
	}

	return client, nil
}

// Send implements the notifier.Notifier interface
func (c *TelegramClient) Send(destination string, notification models.Notification) error {
	if c.bot == nil {
		return fmt.Errorf("telegram client not initialized")
	}

	c.log.Debug().Msg("Sending Telegram notification to chat: " + destination)

	message := formatNotification(notification)

	return c.SendMessage(destination, message, "Markdown")
}

// Name implements the notifier.Notifier interface
func (c *TelegramClient) Name() string {
	return "telegram"
}

// Connect establishes a connection to Telegram and logs bot information
func (c *TelegramClient) Connect() error {
	c.log.Info().
		Str("bot_username", c.bot.Self.UserName).
		Bool("is_bot", c.bot.Self.IsBot).
		Msg("Connected to Telegram")

	return nil
}

// SendMessage sends a message to a Telegram chat
func (c *TelegramClient) SendMessage(chatID string, text string, parseMode string) error {
	// Convert chat ID from string to int64
	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid chat ID: %w", err)
	}

	// Create a new message
	msg := tgbotapi.NewMessage(chatIDInt, text)

	// Set parse mode if provided
	if parseMode != "" {
		msg.ParseMode = parseMode
	}

	// Send the message
	_, err = c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("error sending message to Telegram: %w", err)
	}

	return nil
}

// StartPolling starts polling for updates from Telegram
// This is optional and can be used if you want to receive messages from users
func (c *TelegramClient) StartPolling() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	go func() {
		for update := range updates {
			if update.Message != nil {
				c.log.Info().
					Str("user", update.Message.From.UserName).
					Str("chat_id", fmt.Sprintf("%d", update.Message.Chat.ID)).
					Str("text", update.Message.Text).
					Msg("Received message")

				//TODO: Here you can add command handling if needed
				c.log.Info().Msg("Handling message from user")
			}
		}
	}()
}
