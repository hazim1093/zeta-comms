package notifiers

import (
	"github.com/hazim1093/zeta-comms/pkg/models"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	// Send sends a notification to the specified destination
	Send(destination string, notification models.Notification) error

	// Name returns the name of the notifier (e.g., "slack", "discord")
	Name() string
}
