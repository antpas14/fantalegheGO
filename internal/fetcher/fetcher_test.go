package fetcher

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mocked error")
}

func (errReader) Close() error {
	return nil
}

func TestFetcherImplRetrieveSuccess(t *testing.T) {
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

func TestFetcherImplRetrieveError(t *testing.T) {
	/*// Mock server to simulate an error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send an error response
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))*/
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do not respond to the request, simulating an error
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
	expectedErrorMessage := "Post \"" + server.URL + "/retrieve\": tls: failed to verify certificate: x509: certificate signed by unknown authority"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	// Check that the result is empty
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}

func TestFetcherImplRetrieveHTTPPostError(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection to simulate an error
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	defer server.Close()

	// Create an instance of FetcherImpl with the mock server URL
	fetcher := &FetcherImpl{FetcherUrl: server.URL}

	// Call the Retrieve method with a URL
	result, err := fetcher.Retrieve("http://example.com")

	// Check for an error
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	// Check the error message or any other assertions based on your implementation
	expectedErrorMessage := "HTTP request failed with status code: 503"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	// Check that the result is empty
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}

func TestFetcherImplRetrieveReadBodyError(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send a successful response
		w.Write([]byte(`{"result":"success"}`))
	}))
	defer server.Close()

	// Replace the http.DefaultClient Transport with the mock server's Transport
	http.DefaultClient.Transport = server.Client().Transport

	// Create an instance of FetcherImpl with the mock server URL
	fetcher := &FetcherImpl{FetcherUrl: server.URL}

	// Create a mock response body that returns an error on Read
	mockBody := ioutil.NopCloser(&errReader{})

	// Mock the http.Post function to return the mock response
	http.Post = func(url, contentType string, body io.Reader) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       mockBody,
		}, nil
	}
	defer func() {
		// Restore the original http.Post function after the test
		http.Post = http.DefaultClient.Post
	}()

	// Call the Retrieve method with a URL
	result, err := fetcher.Retrieve("http://example.com")

	// Check for an error
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}

	// Check the error message
	expectedErrorMessage := "mocked error"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	// Check that the result is empty
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}
}
