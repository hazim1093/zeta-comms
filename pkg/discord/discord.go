package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
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
func FormatProposalMessage(title, message, proposalID, status string) *discordgo.MessageEmbed {
	// Create a rich embed for the proposal
	color := getColorForStatus(status)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Proposal Update: %s", title),
		Description: message,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID",
				Value:  proposalID,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  formatStatus(status),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "ZetaChain Governance",
		},
	}

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
