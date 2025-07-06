package calculate

import (
	"errors"
	"fantalegheGO/internal/excel"
	"io"
	"mime/multipart"
	"sort"
	"testing"

	"fantalegheGO/internal/parser"

	api "github.com/antpas14/fantalegheEV-api"
)

// Mocks
type MockFileHeaderOpener struct {
	OpenFunc func() (io.Reader, error)
	FileName string
	FileSize int64
}

func (m *MockFileHeaderOpener) Open() (io.Reader, error) {
	if m.OpenFunc != nil {
		return m.OpenFunc()
	}
	return nil, errors.New("OpenFunc not implemented in MockFileHeaderOpener")
}

func (m *MockFileHeaderOpener) Filename() string {
	return m.FileName
}

func (m *MockFileHeaderOpener) Size() int64 {
	return m.FileSize
}

type MockExcelService struct {
	ReadExcelFunc           func(fileHeader excel.FileHeaderOpener) ([][]string, error)
	ReadExcelFromReaderFunc func(reader io.Reader) ([][]string, error) // Added for the second method
}

func (m *MockExcelService) ReadExcel(fileHeader excel.FileHeaderOpener) ([][]string, error) {
	if m.ReadExcelFunc != nil {
		return m.ReadExcelFunc(fileHeader)
	}
	return nil, errors.New("ReadExcelFunc not implemented in mock")
}

func (m *MockExcelService) ReadExcelFromReader(reader io.Reader) ([][]string, error) {
	if m.ReadExcelFromReaderFunc != nil {
		return m.ReadExcelFromReaderFunc(reader)
	}
	return nil, errors.New("ReadExcelFromReaderFunc not implemented in mock")
}

type MockParser struct {
	GetTeamResultsFunc func(excelRawData [][]string) ([]parser.MatchResults, error)
}

func (m *MockParser) GetTeamResults(excelRawData [][]string) ([]parser.MatchResults, error) {
	if m.GetTeamResultsFunc != nil {
		return m.GetTeamResultsFunc(excelRawData)
	}
	return nil, errors.New("GetTeamResultsFunc not implemented in mock")
}

// --- Test Functions ---

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		name string
		t1   parser.TeamResult
		t2   parser.TeamResult
		want float64
	}{
		{
			name: "Team 1 Wins",
			t1:   parser.TeamResult{Goals: 3},
			t2:   parser.TeamResult{Goals: 1},
			want: 3,
		},
		{
			name: "Team 2 Wins",
			t1:   parser.TeamResult{Goals: 1},
			t2:   parser.TeamResult{Goals: 3},
			want: 0,
		},
		{
			name: "Draw",
			t1:   parser.TeamResult{Goals: 2},
			t2:   parser.TeamResult{Goals: 2},
			want: 1,
		},
		{
			name: "Negative Goals (Edge Case)",
			t1:   parser.TeamResult{Goals: -1},
			t2:   parser.TeamResult{Goals: -5},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculatePoints(tt.t1, tt.t2)
			if got != tt.want {
				t.Errorf("calculatePoints(%v, %v) = %f; want %f", tt.t1, tt.t2, got, tt.want)
			}
		})
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		name    string
		results []parser.MatchResults
		want    []api.Rank
	}{
		{
			name: "Multiple Results",
			results: []parser.MatchResults{
				{
					TeamResults: []parser.TeamResult{
						{Team: "TeamA", Goals: 1, Points: 0},
						{Team: "TeamB", Goals: 2, Points: 3},
						{Team: "TeamC", Goals: 2, Points: 0},
						{Team: "TeamD", Goals: 3, Points: 3},
					},
				},
				{
					TeamResults: []parser.TeamResult{
						{Team: "TeamA", Goals: 3, Points: 3},
						{Team: "TeamB", Goals: 2, Points: 0},
						{Team: "TeamC", Goals: 2, Points: 0},
						{Team: "TeamD", Goals: 3, Points: 3},
					},
				},
			},
			want: []api.Rank{
				{Team: apiString("TeamA"), EvPoints: apiFloat64(2.3333333), Points: apiInt(3)},
				{Team: apiString("TeamB"), EvPoints: apiFloat64(1.6666666), Points: apiInt(3)},
				{Team: apiString("TeamC"), EvPoints: apiFloat64(1.6666666), Points: apiInt(0)},
				{Team: apiString("TeamD"), EvPoints: apiFloat64(5.333333), Points: apiInt(6)},
			},
		},
		{
			name:    "No Results",
			results: []parser.MatchResults{},
			want:    []api.Rank{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculate(tt.results)

			sortRanks(got)
			sortRanks(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("calculate() got %d ranks, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if *got[i].Team != *tt.want[i].Team {
					t.Errorf("Team mismatch at index %d: got %s, want %s", i, *got[i].Team, *tt.want[i].Team)
				}
				if *got[i].Points != *tt.want[i].Points {
					t.Errorf("Points mismatch for %s: got %d, want %d", *got[i].Team, *got[i].Points, *tt.want[i].Points)
				}
				if !floatEquals(*got[i].EvPoints, *tt.want[i].EvPoints, 0.000001) { // Epsilon for float comparison
					t.Errorf("EvPoints mismatch for %s: got %f, want %f", *got[i].Team, *got[i].EvPoints, *tt.want[i].EvPoints)
				}
			}
		})
	}
}

func TestGetRanks(t *testing.T) {
	mockFileHeader := &multipart.FileHeader{
		Filename: "test.xlsx",
		Size:     100,
		Header:   nil,
	}

	tests := []struct {
		name             string
		mockExcelService *MockExcelService
		mockParser       *MockParser
		want             []api.Rank
		wantErr          bool
	}{
		{
			name: "Successful read and parse",
			mockExcelService: &MockExcelService{
				ReadExcelFunc: func(fh excel.FileHeaderOpener) ([][]string, error) {
					return [][]string{{"data"}}, nil // Succeeds
				},
			},
			mockParser: &MockParser{
				GetTeamResultsFunc: func(rawData [][]string) ([]parser.MatchResults, error) {
					return []parser.MatchResults{
						{
							TeamResults: []parser.TeamResult{
								{Team: "TeamA", Goals: 1, Points: 0},
								{Team: "TeamB", Goals: 2, Points: 3},
								{Team: "TeamC", Goals: 2, Points: 0},
								{Team: "TeamD", Goals: 3, Points: 3},
							},
						},
						{
							TeamResults: []parser.TeamResult{
								{Team: "TeamA", Goals: 3, Points: 3},
								{Team: "TeamB", Goals: 2, Points: 0},
								{Team: "TeamC", Goals: 2, Points: 0},
								{Team: "TeamD", Goals: 3, Points: 3},
							},
						},
					}, nil
				},
			},
			want: []api.Rank{
				{Team: apiString("TeamA"), EvPoints: apiFloat64(2.3333333), Points: apiInt(3)},
				{Team: apiString("TeamB"), EvPoints: apiFloat64(1.6666666), Points: apiInt(3)},
				{Team: apiString("TeamC"), EvPoints: apiFloat64(1.6666666), Points: apiInt(0)},
				{Team: apiString("TeamD"), EvPoints: apiFloat64(5.333333), Points: apiInt(6)},
			},
			wantErr: false,
		},
		{
			name: "Excel service ReadExcel error",
			mockExcelService: &MockExcelService{
				ReadExcelFunc: func(fh excel.FileHeaderOpener) ([][]string, error) {
					return [][]string{{"data"}}, errors.New("there is an error") // Succeeds
				},
			},
			mockParser: &MockParser{ // Parser mock won't be called, but needs to be there
				GetTeamResultsFunc: func(rawData [][]string) ([]parser.MatchResults, error) {
					return nil, nil // Should not be called
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Parser GetTeamResults error",
			mockExcelService: &MockExcelService{
				ReadExcelFunc: func(fh excel.FileHeaderOpener) ([][]string, error) {
					return [][]string{{"data"}}, nil // Succeeds
				},
			},
			mockParser: &MockParser{
				GetTeamResultsFunc: func(rawData [][]string) ([]parser.MatchResults, error) {
					return nil, errors.New("parser error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No team results returned by parser",
			mockExcelService: &MockExcelService{
				ReadExcelFunc: func(fh excel.FileHeaderOpener) ([][]string, error) {
					return [][]string{{"data"}}, nil // Succeeds
				},
			},
			mockParser: &MockParser{
				GetTeamResultsFunc: func(rawData [][]string) ([]parser.MatchResults, error) {
					return []parser.MatchResults{}, nil // Empty results
				},
			},
			want:    []api.Rank{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calcImpl := NewCalculateImpl(tt.mockExcelService, tt.mockParser)
			got, err := calcImpl.GetRanks(mockFileHeader)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRanks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				sortRanks(got)
				sortRanks(tt.want)

				// Compare ranks slice content
				if len(got) != len(tt.want) {
					t.Errorf("GetRanks() got %d ranks, want %d ranks", len(got), len(tt.want))
					return
				}
				for i := range got {
					if *got[i].Team != *tt.want[i].Team {
						t.Errorf("Team mismatch at index %d: got %s, want %s", i, *got[i].Team, *tt.want[i].Team)
					}
					if *got[i].Points != *tt.want[i].Points {
						t.Errorf("Points mismatch for %s: got %d, want %d", *got[i].Team, *got[i].Points, *tt.want[i].Points)
					}
					if !floatEquals(*got[i].EvPoints, *tt.want[i].EvPoints, 0.000001) {
						t.Errorf("EvPoints mismatch for %s: got %f, want %f", *got[i].Team, *got[i].EvPoints, *tt.want[i].EvPoints)
					}
				}
			}
		})
	}
}

// Helper functions

func sortRanks(ranks []api.Rank) {
	sort.Slice(ranks, func(i, j int) bool {
		// Primary sort by EvPoints (desc), secondary by Team (asc) for tie-breaking
		if *ranks[i].EvPoints != *ranks[j].EvPoints {
			return *ranks[i].EvPoints > *ranks[j].EvPoints
		}
		return *ranks[i].Team < *ranks[j].Team
	})
}

func floatEquals(a, b, epsilon float64) bool {
	return (a == b) || (a-b < epsilon && b-a < epsilon)
}

func apiString(s string) *string {
	return &s
}

func apiInt(i int) *int {
	return &i
}

func apiFloat64(f float64) *float64 {
	return &f
}
