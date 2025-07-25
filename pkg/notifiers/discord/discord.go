package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
	"github.com/rs/zerolog"
)

type DiscordClient struct {
	log     *zerolog.Logger
	session *discordgo.Session
	botID   string
}

// interface implementation check
var _ notifiers.Notifier = (*DiscordClient)(nil)

func InitializeDiscordClient(logger *zerolog.Logger, botToken string) (*DiscordClient, error) {
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

	if err = client.Connect(); err != nil {
		return nil, err
	}

	client.AddReconnectHandler()

	return client, nil
}

// Send implements the notifier.Notifier interface
func (c *DiscordClient) Send(destination string, notification models.Notification) error {
	if c.session == nil {
		return fmt.Errorf("discord client not initialized")
	}

	c.log.Debug().Msg("Sending Discord notification to channel: " + destination)

	embed := formatNotification(notification)
	content := fmt.Sprintf("New proposal update: %s", notification.Title)

	return c.SendChannelMessage(destination, content, embed)
}

// Name implements the notifier.Notifier interface
func (c *DiscordClient) Name() string {
	return "discord"
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
