package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

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

func NewSlackClient(logger *zerolog.Logger) *SlackClient {
	log := logger.With().Str("service", "slackClient").Logger()

	return &SlackClient{
		log: &log,
	}
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

// FormatProposalMessage creates a formatted Slack message for a proposal
func FormatProposalMessage(title, message, proposalID, status string) Message {
	// Determine color based on status
	color := getColorForStatus(status)

	// Create header section
	headerText := fmt.Sprintf("*Proposal Update:* %s", title)
	headerBlock := Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: headerText,
		},
	}

	// Create details section
	detailsText := fmt.Sprintf("*ID:* %s\n*Status:* %s\n\n%s",
		proposalID, formatStatus(status), message)
	detailsBlock := Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: detailsText,
		},
	}

	// Create blocks array
	blocks := []Block{headerBlock, detailsBlock}

	// Create attachment with color
	attachment := Attachment{
		Color:  color,
		Blocks: blocks,
	}

	// Create message with fallback text and attachment
	return Message{
		Text:        fmt.Sprintf("Proposal %s: %s", proposalID, title),
		Attachments: []Attachment{attachment},
	}
}

// getColorForStatus returns a color hex code based on proposal status
func getColorForStatus(status string) string {
	switch status {
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return "#3AA3E3" // Blue - Action needed
	case "PROPOSAL_STATUS_PASSED":
		return "#2EB886" // Green - Positive outcome
	case "PROPOSAL_STATUS_REJECTED":
		return "#E01E5A" // Red - Negative outcome
	default:
		return "#808080" // Gray - Neutral information
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
