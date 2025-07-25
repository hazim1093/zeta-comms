package zetachain_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/hazim1093/zeta-comms/internal/config"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
)

func TestZetachain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zetachain Suite")
}

const (
	PROPOSAL_STATUS_UNSPECIFIED    = "unspecified"
	PROPOSAL_STATUS_DEPOSIT_PERIOD = "deposit_period"
	PROPOSAL_STATUS_VOTING_PERIOD  = "voting_period"
	PROPOSAL_STATUS_PASSED         = "passed"
	PROPOSAL_STATUS_REJECTED       = "rejected"
	PROPOSAL_STATUS_FAILED         = "failed"
)

var _ = Describe("RESTClient", func() {
	var (
		mockServer   *httptest.Server
		restClient   *zetachain.RESTClient
		testConfig   *config.Config
		mockResponse zetachain.ProposalsResponse
	)

	BeforeEach(func() {
		// Setup mock response
		mockResponse = zetachain.ProposalsResponse{
			Proposals: []zetachain.Proposal{
				{
					ProposalId: "1",
					Status:     PROPOSAL_STATUS_VOTING_PERIOD,
					Title:      "Test Proposal 1",
					Summary:    "This is a test proposal",
					Messages: []zetachain.Message{
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
					Messages: []zetachain.Message{
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
					Messages: []zetachain.Message{
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
	})

	AfterEach(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	Describe("GetProposals", func() {
		Context("when the network exists in config", func() {
			BeforeEach(func() {
				// Create mock server
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify the request path
					Expect(r.URL.Path).To(Equal("/cosmos/gov/v1/proposals"))

					// Verify query parameters
					Expect(r.URL.Query().Get("pagination.limit")).To(Equal("1000"))

					// Return the mock response
					w.Header().Set("Content-Type", "application/json")
					err := json.NewEncoder(w).Encode(mockResponse)
					Expect(err).NotTo(HaveOccurred())
				}))

				// Create test config
				mockURL, _ := url.Parse(mockServer.URL)
				testConfig = &config.Config{
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

				// Create REST client
				restClient = zetachain.NewRESTClient(testConfig, nil)
				restClient.SetRestyClient(resty.New().SetBaseURL(mockServer.URL))
			})

			It("should successfully retrieve proposals", func() {
				response, err := restClient.GetProposals("testnet")
				Expect(err).NotTo(HaveOccurred())
				Expect(response).NotTo(BeNil())
			})

			It("should return proposals with MsgSoftwareUpgrade messages", func() {
				response, err := restClient.GetProposals("testnet")
				Expect(err).NotTo(HaveOccurred())

				// Filter proposals to only include those with MsgSoftwareUpgrade
				var upgradeProposals []zetachain.Proposal
				for _, proposal := range response.Proposals {
					for _, msg := range proposal.Messages {
						if msg.Type == "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade" {
							upgradeProposals = append(upgradeProposals, proposal)
							break
						}
					}
				}

				Expect(upgradeProposals).To(HaveLen(2))
			})

			It("should return proposals in the correct order", func() {
				response, err := restClient.GetProposals("testnet")
				Expect(err).NotTo(HaveOccurred())

				// Filter for upgrade proposals
				var upgradeProposals []zetachain.Proposal
				for _, proposal := range response.Proposals {
					for _, msg := range proposal.Messages {
						if msg.Type == "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade" {
							upgradeProposals = append(upgradeProposals, proposal)
							break
						}
					}
				}

				Expect(upgradeProposals).To(HaveLen(2))
				Expect(upgradeProposals[0].ProposalId).To(Equal("1"))
				Expect(upgradeProposals[0].Status).To(Equal(PROPOSAL_STATUS_VOTING_PERIOD))
				Expect(upgradeProposals[1].ProposalId).To(Equal("3"))
				Expect(upgradeProposals[1].Status).To(Equal(PROPOSAL_STATUS_PASSED))
			})
		})

		Context("when the network does not exist in config", func() {
			BeforeEach(func() {
				testConfig = &config.Config{
					Networks: map[string]struct {
						ApiUrl       url.URL       `mapstructure:"api_url"`
						PollInterval time.Duration `mapstructure:"poll_interval"`
						Audiences    []string      `mapstructure:"audiences"`
					}{},
				}

				restClient = zetachain.NewRESTClient(testConfig, nil)
			})

			It("should return an error", func() {
				response, err := restClient.GetProposals("nonexistent")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("network nonexistent not found in config"))
				Expect(response).To(BeNil())
			})
		})

		Context("when the API request fails", func() {
			BeforeEach(func() {
				// Create mock server that returns 500
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				mockURL, _ := url.Parse(mockServer.URL)
				testConfig = &config.Config{
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

				restClient = zetachain.NewRESTClient(testConfig, nil)
				restClient.SetRestyClient(resty.New().SetBaseURL(mockServer.URL))
			})

			It("should return an error", func() {
				response, err := restClient.GetProposals("testnet")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API request failed with status 500"))
				Expect(response).To(BeNil())
			})
		})

		Context("when the API returns invalid JSON", func() {
			BeforeEach(func() {
				// Create mock server that returns invalid JSON
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte("invalid json"))
				}))

				mockURL, _ := url.Parse(mockServer.URL)
				testConfig = &config.Config{
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

				restClient = zetachain.NewRESTClient(testConfig, nil)
				restClient.SetRestyClient(resty.New().SetBaseURL(mockServer.URL))
			})

			It("should return an error", func() {
				response, err := restClient.GetProposals("testnet")
				Expect(err).To(HaveOccurred())
				Expect(response).To(BeNil())
			})
		})
	})
})
