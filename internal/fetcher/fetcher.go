package fetcher

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"fmt"
)

type Fetcher interface {
	Retrieve(url string) (string, error)
}

type FetcherImpl struct{
    FetcherUrl string
}

func (f *FetcherImpl) Retrieve(url string) (string, error) {
    // Create a JSON request body
    requestBody := map[string]string{
        "url": url,
    }

    // Convert the data to JSON, no err check on marshalling this
    jsonData, _ := json.Marshal(requestBody)
    // Make the POST request
    resp, err := http.Post(f.FetcherUrl + "/retrieve", "application/json", bytes.NewBuffer(jsonData))
    fmt.Printf("pesce")
    if err != nil {
        fmt.Printf("pesce2")
        return "", err
    }
    defer resp.Body.Close()

    // Check the HTTP status code
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
    }

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    // Convert the response body to a string
    responseString := string(body)
    return responseString, nil
}