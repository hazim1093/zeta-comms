package zetachain

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/rs/zerolog"
)

const (
	proposalsPath = "/cosmos/gov/v1/proposals"
)

type RESTClient struct {
	config      *config.Config
	restyClient *resty.Client
	log         *zerolog.Logger
}

type ProposalsResponse struct {
	Proposals  []Proposal `json:"proposals"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

type Proposal struct {
	ProposalId       string      `json:"id"`
	Status           string      `json:"status"`
	Title            string      `json:"title"`
	Summary          string      `json:"summary"`
	Messages         []Message   `json:"messages"`
	FinalTallyResult TallyResult `json:"final_tally_result"`
	SubmitTime       time.Time   `json:"submit_time"`
	DepositEndTime   time.Time   `json:"deposit_end_time"`
	VotingStartTime  time.Time   `json:"voting_start_time"`
	VotingEndTime    time.Time   `json:"voting_end_time"`
	TotalDeposit     []Deposit   `json:"total_deposit"`
	Metadata         string      `json:"metadata"`
	FailedReason     string      `json:"failed_reason,omitempty"`
	Expedited        bool        `json:"expedited"`
}

type TallyResult struct {
	YesCount        string `json:"yes_count"`
	NoCount         string `json:"no_count"`
	AbstainCount    string `json:"abstain_count"`
	NoWithVetoCount string `json:"no_with_veto_count"`
}

type Deposit struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Message struct {
	Type string      `json:"@type"`
	Data MessageData `json:",inline"`
}

type MessageData struct {
	Authority string      `json:"authority,omitempty"`
	Plan      UpgradePlan `json:"plan,omitempty"`
}

type UpgradePlan struct {
	Name   string `json:"name"`
	Height string `json:"height"`
	Info   string `json:"info"`
}

func NewRESTClient(cfg *config.Config, logger *zerolog.Logger) *RESTClient {
	client := resty.New().
		SetHeader("Content-Type", "application/json").
		SetRetryCount(10)

	return &RESTClient{
		config:      cfg,
		restyClient: client,
		log:         logger,
	}
}

func (r *RESTClient) GetProposals(network string) (*ProposalsResponse, error) {
	networkURL, ok := r.config.Networks[network]
	if !ok {
		return nil, fmt.Errorf("network %s not found in config", network)
	}

	var response ProposalsResponse
	resp, err := r.restyClient.R().
		SetResult(&response).
		SetQueryParams(map[string]string{
			"pagination.limit": "1000",
		}).
		Get(networkURL.ApiUrl.String() + proposalsPath)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode())
	}

	return &response, nil
}

// SetRestyClient allows setting a custom resty client for testing purposes
func (r *RESTClient) SetRestyClient(client *resty.Client) {
	r.restyClient = client
}
