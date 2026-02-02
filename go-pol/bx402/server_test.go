package bx402

import (
	"github.com/mark3labs/x402-go"
	x402http "github.com/mark3labs/x402-go/http"
	"net/http"
	"testing"
)

// HONEY on Bepolia 0xFCBD14DC51f0A4d49d5E53C2E0950e0bC26d0Dce

var BeraTestnet = x402.ChainConfig{
	NetworkID:      "bepolia",
	USDCAddress:    "0x406f530cc683E668d74F76d13da1Ec5E8cE582ea",
	Decimals:       6,
	EIP3009Name:    "bxUSDC",
	EIP3009Version: "2",
}

func TestBasicServer(t *testing.T) {
	// Create payment requirement using USDC helper
	requirement, _ := x402.NewUSDCPaymentRequirement(x402.USDCRequirementConfig{
		Chain:            BeraTestnet,
		Amount:           "0.01",                                       // Human-readable USDC amount
		RecipientAddress: "0x160D0E134b78BF083Bd7f03b5d9Fcbcb1c6FF27A", // Bera Main 1
	})

	// Configure middleware
	config := &x402http.Config{
		FacilitatorURL:      "https://x402.testnet.berachain.com/",
		PaymentRequirements: []x402.PaymentRequirement{requirement},
	}

	yourHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	// Protect your endpoint
	middleware := x402http.NewX402Middleware(config)
	http.Handle("/data", middleware(yourHandler))
	http.ListenAndServe(":8080", nil)
}
