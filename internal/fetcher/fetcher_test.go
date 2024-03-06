package fetcher

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

type MockHTTPClient struct {
	StatusCode   int    // Custom status code for the response
	ResponseBody string // Custom response body for the response
	ClientError  bool   // In case mock will return a not nil err
}

func (c *MockHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	if !c.ClientError {
		return &http.Response{
			StatusCode: c.StatusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(c.ResponseBody)), // Simulate response body
		}, nil
	}
	// Simulate returning an error
	return nil, errors.New("mock client error")
}

func TestFetcherImplRetrieveSuccess(t *testing.T) {
	// Define custom status code and response body for the successful scenario
	statusCode := http.StatusOK
	message := `{"result":"success"}`

	// Create an instance of FetcherImpl with the mock HTTP client for success
	fetcher := &FetcherImpl{
		FetcherUrl: "http://example.com",
		Client:     &MockHTTPClient{StatusCode: statusCode, ResponseBody: message, ClientError: false},
	}

	// Call the Retrieve method
	result, err := fetcher.Retrieve("http://example.com")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check the result
	expectedResult := `{"result":"success"}`
	if result != expectedResult {
		t.Errorf("Expected result %s, got %s", expectedResult, result)
	}
}

func TestFetcherImplRetrieveClientError(t *testing.T) {
	// Define custom status code and response body for the successful scenario
	statusCode := http.StatusOK
	message := `{"result":"success"}`

	// Create an instance of FetcherImpl with the mock HTTP client for success
	fetcher := &FetcherImpl{
		FetcherUrl: "http://example.com",
		Client:     &MockHTTPClient{StatusCode: statusCode, ResponseBody: message, ClientError: true},
	}

	// Call the Retrieve method with a URL
	result, err := fetcher.Retrieve("http://example.com")

	// Check for an error
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	// Check the error message
	expectedErrorMessage := "mock client error"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	// Check that the result is empty
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}

func TestFetcherImplRetrieveServerError(t *testing.T) {
	// Define custom status code and response body for the error scenario
	statusCode := http.StatusInternalServerError
	message := "Error"

	// Create an instance of FetcherImpl with the mock HTTP client for failure
	fetcher := &FetcherImpl{
		FetcherUrl: "http://example.com",
		Client:     &MockHTTPClient{StatusCode: statusCode, ResponseBody: message, ClientError: false},
	}

	// Call the Retrieve method with a URL
	result, err := fetcher.Retrieve("http://example.com")

	// Check for an error
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	// Check the error message
	expectedErrorMessage := "HTTP request failed with status code: 500"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	// Check that the result is empty
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}
