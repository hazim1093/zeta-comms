package telegram

import (
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/hazim1093/zeta-comms/pkg/models"
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
func FormatProposalMessage(notification models.Notification) string {
	// Create a formatted message with Markdown
	emoji := getEmojiForStatus(notification.Status)

	formattedMessage := fmt.Sprintf("*[%s]* *Proposal* %s: %s\n\n", notification.Network, notification.ProposalId, notification.Title)
	formattedMessage += fmt.Sprintf("*ID:* %s\n", notification.ProposalId)
	formattedMessage += fmt.Sprintf("*Status:* %s %s\n\n", emoji, formatStatus(notification.Status))

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

				//TODO: Here you can add command handling if needed
				c.log.Info().Msg("Handling message from user")
			}
		}
	}()
}
