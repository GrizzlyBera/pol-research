package go_pol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestAnalyzeMempool(t *testing.T) {

	// url := "https://rpc.berachain.com"
	url := "http://34.47.11.184:8545"

	jsonData := "{\"method\":\"txpool_content\",\"id\":1,\"jsonrpc\":\"2.0\"}"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	require.NoError(t, err)

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Send the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	jsonBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// jsonBytes, err := os.ReadFile("./bera_mempool_1.json")
	// require.NoError(t, err)

	var jsonMap map[string]any
	err = json.Unmarshal(jsonBytes, &jsonMap)
	require.NoError(t, err)

	resultMap, ok := jsonMap["result"].(map[string]any)
	require.True(t, ok)

	pendingMap, ok := resultMap["pending"].(map[string]any)
	require.True(t, ok)
	println(fmt.Sprintf("pending: %v", len(pendingMap)))

	queuedMap, ok := resultMap["queued"].(map[string]any)
	require.True(t, ok)
	println(fmt.Sprintf("queued: %v", len(queuedMap)))
}
