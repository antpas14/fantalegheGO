package calculate

import (
	"fantalegheGO/internal/parser"
	"github.com/stretchr/testify/mock"
	"testing"
	"sync"
	"errors"
)

// MockParser is a mock implementation of the parser.Parser interface
type MockParser struct {
	mock.Mock
}

type MockFetcher struct {
	mock.Mock
}

type Fetcher interface {
	Retrieve(url string) (string, error)
}

func (m *MockFetcher) Retrieve(url string) (string, error) {
	args := m.Called(url)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockParser) GetPoints(rankingsPage string) (map[string]int, error) {
	args := m.Called(rankingsPage)
	return args.Get(0).(map[string]int), args.Error(1)
}

func (m *MockParser) GetResults(calendarPage string) (*sync.Map, error) {
    args := m.Called(calendarPage)
    return args.Get(0).(*sync.Map), args.Error(1)
}

func TestGetRanks(t *testing.T) {
    // Create a mock parser instance
    mockParser := new(MockParser)
    mockFetcher := new(MockFetcher)

    // Define the expected behaviour of the mock parser and fetcher
    results := make(map[int][]parser.TeamResult)
    results[1] = []parser.TeamResult{{Name: "TeamA", Points: 3}, {Name: "TeamB", Points: 1}, {Name: "TeamC", Points: 3}, {Name: "TeamD", Points: 1}}

    mockFetcher.On("Retrieve", mock.Anything).Return("")
    mockParser.On("GetPoints", mock.Anything).Return(map[string]int{"TeamA": 3, "TeamC": 3, "TeamB": 0, "TeamD": 0}, nil)

    mockParser.On("GetResults", mock.Anything).Return(convertToSyncMap(results), nil)

    // Pass the mockParser as an argument to GetRanks
    calculate := &CalculateImpl{}

    calculate.GetRanks("YourLeagueName", mockParser)

    // Assert the results or behavior based on the mockParser's expectations
    points, err := mockParser.GetPoints("")

    t.Log("MockParser GetPoints result (map):", points)
    if err != nil {
        t.Log("MockParser GetPoints result (error):", err)
    }

    results2, err := mockParser.GetResults("")

    t.Log("MockParser GetResults result (map):", results2)
    if err != nil {
        t.Log("MockParser GetResults result (error):", err)
    }
    // Ensure that the mock parser's expectations were met
    mockParser.AssertExpectations(t)
}

func TestGetRanksWithErrorResults(t *testing.T) {
    // Create a mock parser instance
    mockParser := new(MockParser)
    mockFetcher := new(MockFetcher)

    // Define the expected behavior of the mock parser and fetcher
    mockFetcher.On("Retrieve", mock.Anything).Return("")
    mockParser.On("GetPoints", mock.Anything).Return(map[string]int{"TeamA": 3, "TeamC": 3, "TeamB": 0, "TeamD": 0}, nil)

    // Simulating an error in fetching results
    results := make(map[int][]parser.TeamResult)
    mockParser.On("GetResults", mock.Anything).Return(convertToSyncMap(results), errors.New("mock error"))

    calculate := &CalculateImpl{}

    // Call GetRanks and capture the returned error
    _, err := calculate.GetRanks("YourLeagueName", mockParser)

    // Assert that the error is not nil
    if err == nil {
        t.Error("Expected non-nil error, but got nil")
    }
    // Add more assertions as needed based on your specific error handling in GetRanks

    // Ensure that the mock parser's expectations were met
    mockParser.AssertExpectations(t)
}

func TestGetRanksWithErrorPoints(t *testing.T) {
    // Create a mock parser instance
    mockParser := new(MockParser)
    mockFetcher := new(MockFetcher)

    // Define the expected behavior of the mock parser and fetcher
    mockFetcher.On("Retrieve", mock.Anything).Return("")

    // Return an error when GetPoints is called
    mockParser.On("GetPoints", mock.Anything).Return(map[string]int{"TeamA": 3, "TeamC": 3, "TeamB": 0, "TeamD": 0}, errors.New("mock error"))

    // Pass the mockParser as an argument to GetRanks
    calculate := &CalculateImpl{}

    // Call GetRanks and capture the returned error
    _, err := calculate.GetRanks("YourLeagueName", mockParser)

    // Assert that the error is not nil
    if err == nil {
        t.Error("Expected non-nil error, but got nil")
    }
    // Add more assertions as needed based on your specific error handling in GetRanks

    // Ensure that the mock parser's expectations were met
    mockParser.AssertExpectations(t)
}

// Utils
func convertToSyncMap(results map[int][]parser.TeamResult) *sync.Map {
	syncMap := new(sync.Map)

	for key, value := range results {
		syncMap.Store(key, value)
	}

	return syncMap
}
