package models

import (
	"time"

	"github.com/hazim1093/zeta-comms/pkg/zetachain"
)

// Notification represents a formatted notification about a proposal
type Notification struct {
	Network string

	// Core proposal data
	ProposalId string
	Title      string
	Summary    string
	Status     string

	// Software upgrade specific
	UpgradeName  string
	TargetHeight string
	BinaryURLs   map[string]string
	Checksums    map[string]string

	// Voting data
	YesVotes     string
	NoVotes      string
	AbstainVotes string
	VetoVotes    string
	TotalVotes   string

	// Timeline
	SubmitTime    time.Time
	VotingEndTime time.Time

	// Status flags
	Expedited    bool
	FailedReason string

	// Deposit info
	TotalDeposit []zetachain.Deposit
}
