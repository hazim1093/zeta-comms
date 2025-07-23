package zetachain

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hazim1093/zeta-comms/internal/config"
)

const (
	PROPOSAL_STATUS_UNSPECIFIED    = "unspecified"
	PROPOSAL_STATUS_DEPOSIT_PERIOD = "deposit_period"
	PROPOSAL_STATUS_VOTING_PERIOD  = "voting_period"
	PROPOSAL_STATUS_PASSED         = "passed"
	PROPOSAL_STATUS_REJECTED       = "rejected"
	PROPOSAL_STATUS_FAILED         = "failed"
)

// mockRESTClient is a mock implementation of the RESTClient interface
type mockRESTClient struct {
	server      *httptest.Server
	restyClient *resty.Client
}

// NewMockRESTClient creates a new mock REST client
func NewMockRESTClient(server *httptest.Server) *mockRESTClient {
	return &mockRESTClient{
		server:      server,
		restyClient: resty.New().SetHostURL(server.URL),
	}
}

// R returns a resty.Request configured to use the mock server
func (m *mockRESTClient) R() *resty.Request {
	return m.restyClient.R()
}

func TestGetProposals(t *testing.T) {
	// Create a mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path
		if r.URL.Path != "/cosmos/gov/v1/proposals" {
			t.Errorf("Expected request path /cosmos/gov/v1/proposals, got %s", r.URL.Path)
		}

		// Verify query parameters
		if r.URL.Query().Get("pagination.reverse") != "true" {
			t.Errorf("Expected pagination.reverse=true, got %s", r.URL.Query().Get("pagination.reverse"))
		}

		// Create mock response with messages
		mockResponse := ProposalsResponse{
			Proposals: []Proposal{
				{
					ProposalId: "1",
					Status:     PROPOSAL_STATUS_VOTING_PERIOD,
					Title:      "Test Proposal 1",
					Summary:    "This is a test proposal",
					Messages: []Message{
						{
							Type: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
						},
					},
				},
				{
					ProposalId: "2",
					Status:     PROPOSAL_STATUS_PASSED,
					Title:      "Test Proposal 2",
					Summary:    "This is another test proposal",
					Messages: []Message{
						{
							Type: "/cosmos.gov.v1beta1.MsgVote",
						},
					},
				},
				{
					ProposalId: "3",
					Status:     PROPOSAL_STATUS_PASSED,
					Title:      "Test Proposal 3",
					Summary:    "This is a software upgrade proposal",
					Messages: []Message{
						{
							Type: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
						},
					},
				},
			},
			Pagination: struct {
				NextKey string `json:"next_key"`
				Total   string `json:"total"`
			}{
				NextKey: "",
				Total:   "3",
			},
		}

		// Return the mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	// Create a mock client that uses the mock server
	mockClient := NewMockRESTClient(mockServer)

	// Create a test config with the mock server URL
	mockURL, _ := url.Parse(mockServer.URL)
	testConfig := &config.Config{
		Networks: map[string]struct {
			ApiUrl       url.URL       `mapstructure:"api_url"`
			PollInterval time.Duration `mapstructure:"poll_interval"`
			Audiences    []string      `mapstructure:"audiences"`
		}{
			"testnet": {
				ApiUrl:       *mockURL,
				PollInterval: 0,
				Audiences:    nil,
			},
		},
	}

	// Test with testnet network using the mock client and test config
	// Create a RESTClient with the test config
	restClient := &RESTClient{
		config:      testConfig,
		restyClient: mockClient.restyClient,
	}

	response, err := restClient.GetProposals("testnet")
	if err != nil {
		t.Fatalf("Failed to get proposals: %v", err)
	}

	// Validate the response
	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// Should only return proposals with MsgSoftwareUpgrade (proposals 1 and 3)
	if len(response.Proposals) != 2 {
		t.Fatalf("Expected 2 proposals with MsgSoftwareUpgrade, got %d", len(response.Proposals))
	}

	// Check first proposal (should be proposal 1)
	if response.Proposals[0].ProposalId != "1" {
		t.Errorf("Expected proposal ID 1, got %s", response.Proposals[0].ProposalId)
	}

	if response.Proposals[0].Status != PROPOSAL_STATUS_VOTING_PERIOD {
		t.Errorf("Expected status %s, got %s", PROPOSAL_STATUS_VOTING_PERIOD, response.Proposals[0].Status)
	}

	// Check second proposal (should be proposal 3, not 2)
	if response.Proposals[1].ProposalId != "3" {
		t.Errorf("Expected proposal ID 3, got %s", response.Proposals[1].ProposalId)
	}

	if response.Proposals[1].Status != PROPOSAL_STATUS_PASSED {
		t.Errorf("Expected status %s, got %s", PROPOSAL_STATUS_PASSED, response.Proposals[1].Status)
	}

	// Print information for verification
	fmt.Printf("Successfully retrieved %d proposals from mock server\n", len(response.Proposals))
	for i, proposal := range response.Proposals {
		fmt.Printf("Proposal %d: ID=%s, Status=%s, Title=%s\n",
			i+1, proposal.ProposalId, proposal.Status, proposal.Title)
	}
}
