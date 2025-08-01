package pricing

import (
	"context"
	"github.com/calbera/go-pyth-client/examples/query"
	"github.com/calbera/go-pyth-client/hermes"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
	"time"
)

func TestBasicPythAPI(t *testing.T) {

	var testConfig = hermes.Config{
		// Offchain parameters
		APIEndpoint: "https://hermes.pyth.network",
		HTTPTimeout: 1 * time.Second,
		MaxRetries:  2,
		Ecosystem:   "EVM-Stable",

		// Onchain parameters
		UseMock: true, // Uses the mock Pyth contract rather than the real one.
	}

	settings := query.Settings{
		UseEma:           false,
		DesiredPrecision: 0,
		RequestType:      "",
		SingleUpdateFee:  0,
	}

	pythClient, _ := hermes.NewClient(&testConfig, slog.Default())

	// func GetAllLatestPrices(
	//	ctx context.Context, pythClient client.Hermes, qs *Settings,
	//	pairIndexes map[string]uint64, oracleFeeds map[string][]string, uniqueFeeds []string,
	uniqueFeeds := []string{"BERA/USD"}
	pu, err := query.GetAllLatestPrices(context.Background(), pythClient, &settings, nil, nil, uniqueFeeds)
	require.NoError(t, err)
	_ = pu
}
