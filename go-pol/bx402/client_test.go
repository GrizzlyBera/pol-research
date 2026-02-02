package bx402

import (
	"github.com/mark3labs/x402-go"
	x402http "github.com/mark3labs/x402-go/http"
	"github.com/mark3labs/x402-go/signers/evm"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestBasicClient(t *testing.T) {

	// Create USDC token config using helper
	token := x402.NewUSDCTokenConfig(BeraTestnet, 1)

	skHex := os.Getenv("SK_HEX_DEV1")

	// Create signer with your wallet
	signer, err := evm.NewSigner(
		evm.WithPrivateKey(skHex),
		evm.WithNetwork("bepolia"),
		evm.WithToken(token.Address, token.Symbol, token.Decimals),
	)
	require.NoError(t, err)

	// Create client - payments happen automatically
	client, err := x402http.NewClient(x402http.WithSigner(signer))
	require.NoError(t, err)

	resp, err := client.Get("http://localhost:8080/data")
	require.NoError(t, err)
	_ = resp
}
