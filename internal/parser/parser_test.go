package parser

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

// MockDocumentProvider is a mock implementation of DocumentFromReaderProvider for testing purposes.
type MockDocumentProvider struct{}

func (m *MockDocumentProvider) NewDocumentFromReader(r *strings.Reader) (*goquery.Document, error) {
	// Mock implementation returns a dummy document and nil error.
	return nil, errors.New("mocked error")
}

func TestParserImpl_GetPoints(t *testing.T) {
	// Mock HTML content for testing
	mockHTML, err := readHTMLFromFile("../../testdata/ranking.txt")
	if err != nil {
		log.Fatal(err)
	}

	parser := DefaultParserImpl()
	points, err := parser.GetPoints(mockHTML)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPoints := map[string]int{"A": 17, "B": 15}

	for team, expectedPoints := range expectedPoints {
		if points[team] != expectedPoints {
			t.Errorf("Expected points for %s: %d, got: %d", team, expectedPoints, points[team])
		}
	}
}

func TestParserImpl_GetPointsThrowsError(t *testing.T) {
	// Mock HTML content for testing
	mockHTML, err := readHTMLFromFile("../../testdata/ranking.txt")

	parser := CustomParserImpl(&MockDocumentProvider{})
	points, err := parser.GetPoints(mockHTML)

	// Check the error message
	expectedErrorMessage := "mocked error"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	if points != nil {
		t.Fatalf("Points should be nil")
	}

}

func TestParserImpl_GetResults(t *testing.T) {
	// Mock HTML content for testing
	mockHTML := `
		<div class="match-frame">
			<div class="match">
				<div class="team">
					<div class="team-name">TeamA</div>
					<div class="team-fpt">5.0</div>
					<div class="team-score">3</div>
				</div>
				<div class="team">
					<div class="team-name">TeamB</div>
					<div class="team-fpt">4.0</div>
					<div class="team-score">2</div>
				</div>
			</div>
		</div>
	`

	parser := DefaultParserImpl()
	results, err := parser.GetResults(mockHTML)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedResults := map[string][]TeamResult{
		"1": {{Name: "TeamA", Points: 3}, {Name: "TeamB", Points: 2}},
	}

	for key, expectedTeamResults := range expectedResults {
		result, ok := results.Load(key)
		if !ok {
			t.Errorf("Expected result for key %s not found", key)
			continue
		}

		teamResults, ok := result.([]TeamResult)
		if !ok {
			t.Errorf("Invalid result type for key %s", key)
			continue
		}

		for i, expectedTeamResult := range expectedTeamResults {
			if i >= len(teamResults) {
				t.Errorf("Expected more team results for key %s", key)
				break
			}

			if teamResults[i] != expectedTeamResult {
				t.Errorf("Expected team result for key %s: %v, got: %v", key, expectedTeamResult, teamResults[i])
			}
		}
	}
}

func TestParserImpl_GetResultsInvalidHTML(t *testing.T) {
	// Mock HTML content for testing
	mockHTML := ``

	parser := CustomParserImpl(&MockDocumentProvider{})
	results, err := parser.GetResults(mockHTML)

	// Check the error message
	expectedErrorMessage := "mocked error"

	if err.Error() != expectedErrorMessage {
		t.Errorf("Expected error message %s, got %v", expectedErrorMessage, err)
	}

	isEmpty := true
	results.Range(func(key, value interface{}) bool {
		return false
	})

	if !isEmpty {
		t.Fatalf("Results should be empty")
	}
}

// Utils functions

func readHTMLFromFile(filePath string) (string, error) {
	// Read the content of the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string
	htmlContent := string(content)
	return htmlContent, nil
}
