package fetcher

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var urlFetcher = "http://localhost:5000/retrieve"

func Retrieve(url string) (string, error) {
	// Create a JSON request body
	requestBody := map[string]string{
		"url": url,
	}

	// Convert the data to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Make the POST request
	resp, err := http.Post(urlFetcher, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Convert the response body to a string
	responseString := string(body)
	return responseString, nil
}