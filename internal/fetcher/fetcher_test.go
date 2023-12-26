package fetcher

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetcherImpl_Retrieve_Success(t *testing.T) {
	// Mock server to simulate a successful response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Check the request URL
		if r.URL.Path != "/retrieve" {
			t.Errorf("Expected request to /retrieve, got %s", r.URL.Path)
		}

		// Read the request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		// Verify the request body
		expectedBody := `{"url":"http://example.com"}`
		if string(body) != expectedBody {
			t.Errorf("Expected request body %s, got %s", expectedBody, string(body))
		}

		// Send a successful response
		w.Write([]byte(`{"result":"success"}`))
	}))

	defer server.Close()

	// Create an instance of FetcherImpl with the mock server URL
	fetcher := &FetcherImpl{FetcherUrl: server.URL}

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

func TestFetcherImpl_Retrieve_Error(t *testing.T) {
	// Mock server to simulate an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send an error response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))

	defer server.Close()

	// Create an instance of FetcherImpl with the mock server URL
	fetcher := &FetcherImpl{FetcherUrl: server.URL}

	// Call the Retrieve method
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
