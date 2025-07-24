package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hazim1093/zeta-comms/pkg/models"
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
func FormatProposalMessage(notification models.Notification) Message {
	// Determine color based on status
	color := getColorForStatus(notification.Status)

	// Create header section
	headerText := fmt.Sprintf("*[%s]* *Proposal* %s: %s", notification.Network, notification.ProposalId, notification.Title)
	headerBlock := Block{
		Type: "section",
		Text: &Text{
			Type: "mrkdwn",
			Text: headerText,
		},
	}

	// Build message content
	var messageContent string

	// Add software upgrade info if available
	if notification.UpgradeName != "" {
		messageContent += fmt.Sprintf("*Upgrade:* %s\n*Target Height:* %s\n",
			notification.UpgradeName, notification.TargetHeight)
	}

	// Add deposit information if available
	if len(notification.TotalDeposit) > 0 {
		messageContent += "*Deposits:*\n"
		for _, deposit := range notification.TotalDeposit {
			messageContent += fmt.Sprintf("• %s %s\n", deposit.Amount, deposit.Denom)
		}
		messageContent += "\n"
	}

	// Add voting results using the common formatter
	// Format the voting results section
	messageContent = "*Voting Results:*\n"

	if notification.TotalVotes != "" {
		messageContent += fmt.Sprintf("• Yes: %s\n", notification.YesVotes)
		messageContent += fmt.Sprintf("• No: %s\n", notification.NoVotes)
		messageContent += fmt.Sprintf("• Abstain: %s\n", notification.AbstainVotes)
		messageContent += fmt.Sprintf("• Veto: %s\n", notification.VetoVotes)
		messageContent += "\n"
		messageContent += "*Total Votes:* " + notification.TotalVotes + "\n"
	} else {
		messageContent += "No voting results available.\n"
	}

	messageContent += "\n"

	// Add timeline information
	if !notification.SubmitTime.IsZero() {
		messageContent += fmt.Sprintf("*Submitted:* %s\n", notification.SubmitTime.Format(time.RFC1123))
	}
	if !notification.VotingEndTime.IsZero() {
		messageContent += fmt.Sprintf("*Voting Ends:* %s\n", notification.VotingEndTime.Format(time.RFC1123))
	}

	// Add expedited flag if true
	if notification.Expedited {
		messageContent += "*Expedited:* Yes\n"
	}

	// Add failed reason if available
	if notification.FailedReason != "" {
		messageContent += fmt.Sprintf("*Failed Reason:* %s\n", notification.FailedReason)
	}

	// Add summary with a separator
	messageContent += "\n*Summary:*\n" + notification.Summary

	// Create details section
	detailsText := fmt.Sprintf("*ID:* %s\n*Status:* %s\n\n%s",
		notification.ProposalId, formatStatus(notification.Status), messageContent)
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
		Text:        fmt.Sprintf("New proposal notification for %s", notification.Network),
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
