package telegram

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

// TelegramClient handles communication with Telegram API
type TelegramClient struct {
	log *zerolog.Logger
	bot *tgbotapi.BotAPI
}

// NewTelegramClient creates a new Telegram client
func NewTelegramClient(logger *zerolog.Logger, botToken string) (*TelegramClient, error) {
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

	return client, nil
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

// FormatProposalMessage creates a formatted Telegram message for a proposal
func FormatProposalMessage(title, message, proposalID, status string) string {
	// Create a formatted message with Markdown
	emoji := getEmojiForStatus(status)

	formattedMessage := fmt.Sprintf("*Proposal Update: %s*\n\n", title)
	formattedMessage += fmt.Sprintf("*ID:* %s\n", proposalID)
	formattedMessage += fmt.Sprintf("*Status:* %s %s\n\n", emoji, formatStatus(status))
	formattedMessage += message
	formattedMessage += fmt.Sprintf("\n\n_Updated at: %s_", time.Now().Format(time.RFC1123))

	return formattedMessage
}

// getEmojiForStatus returns an emoji based on proposal status
func getEmojiForStatus(status string) string {
	switch status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return "🗳️"
	case "PROPOSAL_STATUS_PASSED":
		return "✅"
	case "PROPOSAL_STATUS_REJECTED":
		return "❌"
	default:
		return "ℹ️"
	}
}

// formatStatus returns a human-readable version of the proposal status
func formatStatus(status string) string {
	switch status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return "Voting Period"
	case "PROPOSAL_STATUS_PASSED":
		return "Passed"
	case "PROPOSAL_STATUS_REJECTED":
		return "Rejected"
	default:
		return status
	}
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

				// Here you can add command handling if needed
			}
		}
	}()
}
