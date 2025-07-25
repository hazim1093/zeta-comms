package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hazim1093/zeta-comms/pkg/models"
	"github.com/hazim1093/zeta-comms/pkg/notifiers"
	"github.com/rs/zerolog"
)

// Message represents a Slack message with blocks for rich formatting
type Message struct {
	Text        string       `json:"text"`
	Blocks      []Block      `json:"blocks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Block represents a Slack block element
type Block struct {
	Type string `json:"type"`
	Text *Text  `json:"text,omitempty"`
}

// Text represents text content within a Slack block
type Text struct {
	Type string `json:"type"` // plain_text or mrkdwn
	Text string `json:"text"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color  string  `json:"color,omitempty"` // For color-coding based on status
	Blocks []Block `json:"blocks,omitempty"`
}

type SlackClient struct {
	log *zerolog.Logger
}

// Ensure SlackClient implements the notifier.Notifier interface
var _ notifiers.Notifier = (*SlackClient)(nil)

func NewSlackClient(logger *zerolog.Logger) *SlackClient {
	log := logger.With().Str("service", "slackClient").Logger()

	return &SlackClient{
		log: &log,
	}
}

// Send implements the notifier.Notifier interface
func (c *SlackClient) Send(destination string, notification models.Notification) error {
	c.log.Debug().Msg("Sending Slack notification to webhook")

	message := formatNotification(notification)

	return c.SendWebhookMessage(destination, message)
}

// Name implements the notifier.Notifier interface
func (c *SlackClient) Name() string {
	return "slack"
}

// SendWebhookMessage sends a message to a Slack webhook URL
func (c *SlackClient) SendWebhookMessage(webhookURL string, message Message) error {
	// Validate webhook URL
	_, err := url.ParseRequestURI(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send HTTP POST request to webhook URL
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send message to Slack: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
