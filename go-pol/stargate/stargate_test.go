package stargate

import (
	go_pol "github.com/GrizzlyBera/pol-research/go-pol"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStargateBasics(t *testing.T) {

	reqParams := map[string]string{
		"srcToken":    "0x688e72142674041f8f6Af4c808a4045cA1D6aC82", // BYUSD
		"srcChainKey": "bera",
		"dstToken":    "0x6c3ea9036406852006290770BEdFcAbA0e23A0e8", // PYUSD on ETH
		// "dstToken":     "0x46850aD61C2B7d64d08c9C754F45254596696984", // PYUSD on Arb
		"dstChainKey":  "ethereum",
		"srcAddress":   "0x0CEbC523D6399DF26Bec5724851677bA1cE723A5",
		"dstAddress":   "0x0CEbC523D6399DF26Bec5724851677bA1cE723A5",
		"srcAmount":    "1000000",
		"dstAmountMin": "950000",
	}

	respBytes, err := go_pol.GetJsonHttpGet("https://stargate.finance/api/v1/quotes", reqParams)
	require.NoError(t, err)

	println(string(respBytes))
}
