package go_pol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func GetJsonHttpGet(baseUrl string, reqParams map[string]string) (respBytes []byte, err error) {

	params := url.Values{}
	for k, v := range reqParams {
		params.Set(k, v)
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	// Assign the encoded query parameters to the URL's RawQuery field
	u.RawQuery = params.Encode()

	// Perform the GET request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // Ensure the response body is closed

	if resp.StatusCode == 200 {
		return io.ReadAll(resp.Body)
	} else {
		return nil, fmt.Errorf("http status error code: %d", resp.StatusCode)
	}
}

func GetJsonHttpPost(reqData any) (respBytes []byte, err error) {

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", "https://api.relay.link/quote", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Create a custom HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return io.ReadAll(resp.Body)
	} else {
		return nil, fmt.Errorf("http status error code: %d", resp.StatusCode)
	}
}
