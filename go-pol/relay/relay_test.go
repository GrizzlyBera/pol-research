package relay

import (
	"fmt"
	go_pol "github.com/GrizzlyBera/pol-research/go-pol"
	"github.com/stretchr/testify/require"
	"testing"
)

type BaseQuoteRequest struct {
	User                string `json:"user"`
	OriginChainId       uint64 `json:"originChainId"`
	OriginCurrency      string `json:"originCurrency"`
	DestinationChainId  uint64 `json:"destinationChainId"`
	DestinationCurrency string `json:"destinationCurrency"`
	Amount              string `json:"amount"`
	TradeType           string `json:"tradeType"`
}

type ExtendedQuoteRequest struct {
	BaseQuoteRequest
	SubsidizeFees          bool   `json:"subsidizeFees"`
	MaxSubsidizationAmount string `json:"maxSubsidizationAmount"`
}

func TestRelayBasics(t *testing.T) {

	quote := ExtendedQuoteRequest{
		BaseQuoteRequest: BaseQuoteRequest{
			User:                "0x0CEbC523D6399DF26Bec5724851677bA1cE723A5",
			OriginChainId:       80094,                                        // Berachain
			OriginCurrency:      "0xFCBD14DC51f0A4d49d5E53C2E0950e0bC26d0Dce", // HONEY
			DestinationChainId:  42161,                                        // Arb One
			DestinationCurrency: "0x46850aD61C2B7d64d08c9C754F45254596696984", // PYUSD
			Amount:              "10000000",
			TradeType:           "EXACT_OUTPUT",
		},
		SubsidizeFees:          true,
		MaxSubsidizationAmount: "1000000",
	}

	respBytes, err := go_pol.GetJsonHttpPost(quote)
	require.NoError(t, err)

	fmt.Println("Response JSON:", string(respBytes))
}
