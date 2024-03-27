package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

type Fetcher interface {
	Retrieve(url string) (string, error)
}

type FetcherImpl struct {
	FetcherUrl string
	Client     HTTPClient // Use the HTTPClient interface for HTTP requests
}

func (f *FetcherImpl) Retrieve(url string) (string, error) {
	// Create a JSON request body
	requestBody := map[string]string{
		"url": url,
	}

	// Convert the data to JSON, no err check on marshalling this
	jsonData, _ := json.Marshal(requestBody)
	// Make the POST request
	resp, err := f.Client.Post(f.FetcherUrl+"/retrieve", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, _ := io.ReadAll(resp.Body)

	// Convert the response body to a string
	responseString := string(body)
	return responseString, nil
}
