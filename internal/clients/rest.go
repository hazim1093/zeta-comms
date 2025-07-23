package clients

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
	ProposalId string    `json:"id"`
	Status     string    `json:"status"`
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	Messages   []Message `json:"messages"`
}

type Message struct {
	Type string    `json:"@type"`
	Name string    `json:"name,omitempty"`
	Time time.Time `json:"time,omitempty"`
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
	// Get network URL
	networkURL, ok := r.config.Networks[network]
	if !ok {
		return nil, fmt.Errorf("network %s not found in config", network)
	}

	// Make request to proposals endpoint with pagination.reverse=true
	var response ProposalsResponse
	resp, err := r.restyClient.R().
		SetResult(&response).
		SetQueryParams(map[string]string{
			"pagination.reverse": "true",
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
