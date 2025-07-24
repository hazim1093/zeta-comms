package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/rs/zerolog"
)

// DiscordClient handles communication with Discord API
type DiscordClient struct {
	log     *zerolog.Logger
	session *discordgo.Session
	botID   string
}

// NewDiscordClient creates a new Discord client
func NewDiscordClient(logger *zerolog.Logger, botToken string) (*DiscordClient, error) {
	log := logger.With().Str("service", "discordClient").Logger()

	// Create a new Discord session using the bot token
	session, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	client := &DiscordClient{
		log:     &log,
		session: session,
	}

	return client, nil
}

// Connect establishes a connection to Discord and gets the bot ID
func (c *DiscordClient) Connect() error {
	// Open a websocket connection to Discord
	err := c.session.Open()
	if err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}

	// Get bot ID
	user, err := c.session.User("@me")
	if err != nil {
		c.session.Close()
		return fmt.Errorf("error getting bot user: %w", err)
	}
	c.botID = user.ID

	c.log.Info().Str("bot_id", c.botID).Msg("Connected to Discord")
	return nil
}

// Close closes the Discord connection
func (c *DiscordClient) Close() error {
	c.log.Info().Msg("Closing Discord connection")
	return c.session.Close()
}

// SendChannelMessage sends a message to a Discord channel
func (c *DiscordClient) SendChannelMessage(channelID string, content string, embed *discordgo.MessageEmbed) error {
	_, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: content,
		Embeds:  []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		return fmt.Errorf("error sending message to Discord channel: %w", err)
	}
	return nil
}

// FormatProposalMessage creates a formatted Discord embed for a proposal
func FormatProposalMessage(notification models.Notification) *discordgo.MessageEmbed {
	// Create a rich embed for the proposal
	color := getColorForStatus(notification.Status)

	title := fmt.Sprintf("[%s] Proposal #%s: %s", notification.Network, notification.ProposalId, notification.Title)
	embed := &discordgo.MessageEmbed{
		Title:     title,
		Color:     color,
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "ZetaChain Governance",
		},
	}

	// Build description
	var description string

	description = fmt.Sprintf("**ID:** %s\n**Status:** %s\n\n", notification.ProposalId, formatStatus(notification.Status))

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

	description += "*Voting Results:*\n"
	if notification.TotalVotes != "" {
		description += fmt.Sprintf("• Yes: %s\n", notification.YesVotes)
		description += fmt.Sprintf("• No: %s\n", notification.NoVotes)
		description += fmt.Sprintf("• Abstain: %s\n", notification.AbstainVotes)
		description += fmt.Sprintf("• Veto: %s\n", notification.VetoVotes)
		description += "\n"
		description += "*Total Votes:* " + notification.TotalVotes + "\n"
	} else {
		description += "No voting results available.\n"
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

// formatStatus returns a human-readable version of the proposal status
func formatStatus(status string) string {
	switch status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return "🗳️ Voting Period"
	case "PROPOSAL_STATUS_PASSED":
		return "✅ Passed"
	case "PROPOSAL_STATUS_REJECTED":
		return "❌ Rejected"
	default:
		return status
	}
}

// AddReconnectHandler adds a handler to automatically reconnect if the connection is lost
func (c *DiscordClient) AddReconnectHandler() {
	c.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		c.log.Info().Msg("Discord connection established")
	})

	c.session.AddHandler(func(s *discordgo.Session, r *discordgo.Resumed) {
		c.log.Info().Msg("Discord connection resumed")
	})

	c.session.AddHandler(func(s *discordgo.Session, r *discordgo.Disconnect) {
		c.log.Warn().Msg("Discord connection lost, will automatically reconnect")
	})
}
