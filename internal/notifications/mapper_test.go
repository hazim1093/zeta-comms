package notifications_test

import (
	"testing"
	"time"

	"github.com/hazim1093/zeta-comms/internal/notifications"
	"github.com/hazim1093/zeta-comms/pkg/zetachain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNotifications(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Notifications Suite")
}

var _ = Describe("Mapper", func() {
	var (
		network  string
		proposal zetachain.Proposal
	)

	BeforeEach(func() {
		network = "testnet"
		proposal = zetachain.Proposal{
			ProposalId: "123",
			Status:     "voting_period",
			Title:      "Test Proposal",
			Summary:    "This is a test proposal for testing purposes",
			Messages:   []zetachain.Message{},
			FinalTallyResult: zetachain.TallyResult{
				YesCount:        "1000000000000000000", // 1 ZETA in azeta
				NoCount:         "500000000000000000",  // 0.5 ZETA in azeta
				AbstainCount:    "200000000000000000",  // 0.2 ZETA in azeta
				NoWithVetoCount: "100000000000000000",  // 0.1 ZETA in azeta
			},
			SubmitTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			VotingEndTime: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			TotalDeposit: []zetachain.Deposit{
				{
					Denom:  "azeta",
					Amount: "5000000000000000000", // 5 ZETA
				},
			},
			Expedited: false,
		}
	})

	Describe("MapFromProposal", func() {
		Context("with basic proposal data", func() {
			It("should correctly map basic fields", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.Network).To(Equal("testnet"))
				Expect(result.ProposalId).To(Equal("123"))
				Expect(result.Title).To(Equal("Test Proposal"))
				Expect(result.Summary).To(Equal("This is a test proposal for testing purposes"))
				Expect(result.Status).To(Equal("voting_period"))
			})

			It("should correctly map timeline fields", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.SubmitTime).To(Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
				Expect(result.VotingEndTime).To(Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)))
			})

			It("should correctly map status flags", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.Expedited).To(BeFalse())
				Expect(result.FailedReason).To(BeEmpty())
			})
		})

		Context("with upgrade information", func() {
			BeforeEach(func() {
				proposal.Messages = []zetachain.Message{
					{
						Type: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
						Data: zetachain.MessageData{
							Plan: zetachain.UpgradePlan{
								Name:   "v1.2.0-upgrade",
								Height: "1234567",
								Info:   "Upgrade to v1.2.0",
							},
						},
					},
				}
			})

			It("should extract upgrade name and target height", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.UpgradeName).To(Equal("v1.2.0-upgrade"))
				Expect(result.TargetHeight).To(Equal("1234567"))
			})
		})

		Context("without upgrade information", func() {
			BeforeEach(func() {
				proposal.Messages = []zetachain.Message{
					{
						Type: "/cosmos.gov.v1beta1.MsgVote",
						Data: zetachain.MessageData{},
					},
				}
			})

			It("should have empty upgrade fields", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.UpgradeName).To(BeEmpty())
				Expect(result.TargetHeight).To(BeEmpty())
			})
		})

		Context("with multiple messages including upgrade", func() {
			BeforeEach(func() {
				proposal.Messages = []zetachain.Message{
					{
						Type: "/cosmos.gov.v1beta1.MsgVote",
						Data: zetachain.MessageData{},
					},
					{
						Type: "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
						Data: zetachain.MessageData{
							Plan: zetachain.UpgradePlan{
								Name:   "multi-message-upgrade",
								Height: "7654321",
							},
						},
					},
				}
			})

			It("should extract upgrade information from the correct message", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.UpgradeName).To(Equal("multi-message-upgrade"))
				Expect(result.TargetHeight).To(Equal("7654321"))
			})
		})

		Context("with vote calculations", func() {
			It("should correctly calculate vote percentages", func() {
				result := notifications.MapFromProposal(network, proposal)

				// Total votes: 1.8 ZETA = 0.0000018M
				// Yes: 1.0/1.8 = 55.56%
				// No: 0.5/1.8 = 27.78%
				// Abstain: 0.2/1.8 = 11.11%
				// Veto: 0.1/1.8 = 5.56%

				Expect(result.YesVotes).To(Equal("0.000M (55.56%)"))
				Expect(result.NoVotes).To(Equal("0.000M (27.78%)"))
				Expect(result.AbstainVotes).To(Equal("0.000M (11.11%)"))
				Expect(result.VetoVotes).To(Equal("0.000M (5.56%)"))
				Expect(result.TotalVotes).To(Equal("0.000M"))
			})
		})

		Context("with zero votes", func() {
			BeforeEach(func() {
				proposal.FinalTallyResult = zetachain.TallyResult{
					YesCount:        "0",
					NoCount:         "0",
					AbstainCount:    "0",
					NoWithVetoCount: "0",
				}
			})

			It("should handle zero votes gracefully", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.YesVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.NoVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.AbstainVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.VetoVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.TotalVotes).To(Equal("0.000M"))
			})
		})

		Context("with very large vote counts", func() {
			BeforeEach(func() {
				proposal.FinalTallyResult = zetachain.TallyResult{
					YesCount:        "1000000000000000000000", // 1000 ZETA
					NoCount:         "500000000000000000000",  // 500 ZETA
					AbstainCount:    "200000000000000000000",  // 200 ZETA
					NoWithVetoCount: "100000000000000000000",  // 100 ZETA
				}
			})

			It("should correctly format large vote counts", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.YesVotes).To(Equal("0.001M (55.56%)"))
				Expect(result.NoVotes).To(Equal("0.001M (27.78%)"))
				Expect(result.TotalVotes).To(Equal("0.002M"))
			})
		})

		Context("with deposit conversion", func() {
			It("should convert azeta deposits to ZETA", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.TotalDeposit).To(HaveLen(1))
				Expect(result.TotalDeposit[0].Denom).To(Equal("ZETA"))
				Expect(result.TotalDeposit[0].Amount).To(Equal("5.00"))
			})
		})

		Context("with mixed deposit denominations", func() {
			BeforeEach(func() {
				proposal.TotalDeposit = []zetachain.Deposit{
					{
						Denom:  "azeta",
						Amount: "1000000000000000000", // 1 ZETA
					},
					{
						Denom:  "uatom",
						Amount: "5000000", // 5 ATOM
					},
					{
						Denom:  "azeta",
						Amount: "2500000000000000000", // 2.5 ZETA
					},
				}
			})

			It("should convert only azeta deposits and keep others as-is", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.TotalDeposit).To(HaveLen(3))
				Expect(result.TotalDeposit[0].Denom).To(Equal("ZETA"))
				Expect(result.TotalDeposit[0].Amount).To(Equal("1.00"))
				Expect(result.TotalDeposit[1].Denom).To(Equal("uatom"))
				Expect(result.TotalDeposit[1].Amount).To(Equal("5000000"))
				Expect(result.TotalDeposit[2].Denom).To(Equal("ZETA"))
				Expect(result.TotalDeposit[2].Amount).To(Equal("2.50"))
			})
		})

		Context("with empty deposit", func() {
			BeforeEach(func() {
				proposal.TotalDeposit = []zetachain.Deposit{}
			})

			It("should handle empty deposit slice", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.TotalDeposit).To(BeEmpty())
			})
		})

		Context("with failed proposal", func() {
			BeforeEach(func() {
				proposal.Status = "failed"
				proposal.FailedReason = "insufficient deposit"
			})

			It("should include failed reason", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.Status).To(Equal("failed"))
				Expect(result.FailedReason).To(Equal("insufficient deposit"))
			})
		})

		Context("with expedited proposal", func() {
			BeforeEach(func() {
				proposal.Expedited = true
			})

			It("should mark as expedited", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.Expedited).To(BeTrue())
			})
		})

		Context("with invalid vote count strings", func() {
			BeforeEach(func() {
				proposal.FinalTallyResult = zetachain.TallyResult{
					YesCount:        "invalid",
					NoCount:         "not_a_number",
					AbstainCount:    "also_invalid",
					NoWithVetoCount: "still_invalid",
				}
			})

			It("should handle invalid vote counts gracefully", func() {
				result := notifications.MapFromProposal(network, proposal)

				// All should be zero since parsing fails
				Expect(result.YesVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.NoVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.AbstainVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.VetoVotes).To(Equal("0.000M (0.00%)"))
				Expect(result.TotalVotes).To(Equal("0.000M"))
			})
		})

		Context("with invalid deposit amounts", func() {
			BeforeEach(func() {
				proposal.TotalDeposit = []zetachain.Deposit{
					{
						Denom:  "azeta",
						Amount: "invalid_amount",
					},
				}
			})

			It("should handle invalid deposit amounts gracefully", func() {
				result := notifications.MapFromProposal(network, proposal)

				Expect(result.TotalDeposit).To(HaveLen(1))
				Expect(result.TotalDeposit[0].Denom).To(Equal("ZETA"))
				Expect(result.TotalDeposit[0].Amount).To(Equal("0.00"))
			})
		})
	})
})
