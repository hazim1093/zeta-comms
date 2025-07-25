package notifiers

// FormatStatus returns a human-readable version of the proposal status
func FormatStatus(status string) string {
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
