package parser

import (
	"reflect"
	"testing"
)

// --- Test Functions ---

func TestCalculateMatchPoints(t *testing.T) {
	tests := []struct {
		name       string
		ourGoals   int
		theirGoals int
		want       int
	}{
		{
			name:       "Win",
			ourGoals:   2,
			theirGoals: 1,
			want:       3,
		},
		{
			name:       "Loss",
			ourGoals:   0,
			theirGoals: 1,
			want:       0,
		},
		{
			name:       "Draw",
			ourGoals:   1,
			theirGoals: 1,
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMatchPoints(tt.ourGoals, tt.theirGoals)
			if got != tt.want {
				t.Errorf("calculateMatchPoints(%d, %d) = %d; want %d", tt.ourGoals, tt.theirGoals, got, tt.want)
			}
		})
	}
}

func TestGetTeamResult(t *testing.T) {
	tests := []struct {
		name  string
		match []string
		want  []TeamResult
	}{
		{
			name:  "Valid Match Row",
			match: []string{"TeamA", "P1", "G1", "TeamB", "2-1", "P2", "G2", "TeamC", "P3", "G3"}, // Only first 5 elements matter for getTeamResult
			want: []TeamResult{
				{Team: "TeamA", Goals: 2, Points: 3},
				{Team: "TeamB", Goals: 1, Points: 0},
			},
		},
		{
			name:  "Draw Match Row",
			match: []string{"TeamX", "P1", "G1", "TeamY", "0-0", "P2", "G2", "TeamZ", "P3", "G3"},
			want: []TeamResult{
				{Team: "TeamX", Goals: 0, Points: 1},
				{Team: "TeamY", Goals: 0, Points: 1},
			},
		},
		{
			name:  "Loss Match Row",
			match: []string{"TeamM", "P1", "G1", "TeamN", "1-3", "P2", "G2", "TeamO", "P3", "G3"},
			want: []TeamResult{
				{Team: "TeamM", Goals: 1, Points: 0},
				{Team: "TeamN", Goals: 3, Points: 3},
			},
		},
		{
			name:  "Insufficient Columns",
			match: []string{"TeamA", "P1", "G1", "TeamB"},
			want:  []TeamResult{},
		},
		{
			name:  "Invalid Goals Format (No Dash)",
			match: []string{"TeamA", "P1", "G1", "TeamB", "21"},
			want:  []TeamResult{},
		},
		{
			name:  "Invalid Goals Format (Non-Numeric)",
			match: []string{"TeamA", "P1", "G1", "TeamB", "a-b"},
			want:  []TeamResult{},
		},
		{
			name:  "Empty Goal String",
			match: []string{"TeamA", "P1", "G1", "TeamB", "-"},
			want:  []TeamResult{},
		},
		{
			name:  "Empty First Goal String",
			match: []string{"TeamA", "P1", "G1", "TeamB", "-1"},
			want:  []TeamResult{},
		},
		{
			name:  "Empty Second Goal String",
			match: []string{"TeamA", "P1", "G1", "TeamB", "1-"},
			want:  []TeamResult{},
		},
		{
			name:  "Empty Match Slice",
			match: []string{},
			want:  []TeamResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTeamResult(tt.match)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTeamResult(%v) = %v; want %v", tt.match, got, tt.want)
			}
		})
	}
}

func TestSplitRows(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want [][]string
	}{
		{
			name: "Standard 10-Column Rows",
			rows: [][]string{
				{"A1", "A2", "A3", "A4", "A5", "B1", "B2", "B3", "B4", "B5"},
				{"C1", "C2", "C3", "C4", "C5", "D1", "D2", "D3", "D4", "D5"},
			},
			want: [][]string{
				{"A1", "A2", "A3", "A4", "A5"},
				{"C1", "C2", "C3", "C4", "C5"},
				{"B1", "B2", "B3", "B4", "B5"},
				{"D1", "D2", "D3", "D4", "D5"},
			},
		},
		{
			name: "Rows with Incorrect Column Count",
			rows: [][]string{
				{"A1", "A2", "A3", "A4", "A5", "B1", "B2", "B3", "B4", "B5"},
				{"C1", "C2", "C3", "C4", "C5"}, // This row should be skipped
				{"E1", "E2", "E3", "E4", "E5", "F1", "F2", "F3", "F4", "F5"},
			},
			want: [][]string{
				{"A1", "A2", "A3", "A4", "A5"},
				{"E1", "E2", "E3", "E4", "E5"},
				{"B1", "B2", "B3", "B4", "B5"},
				{"F1", "F2", "F3", "F4", "F5"},
			},
		},
		{
			name: "Empty Input",
			rows: [][]string{},
			want: nil,
		},
		{
			name: "All Rows Incorrect Length",
			rows: [][]string{
				{"A1", "A2"},
				{"B1", "B2", "B3"},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitRows(tt.rows)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitRows(%v) = %v; want %v", tt.rows, got, tt.want)
			}
		})
	}
}

func TestGetTeamResults(t *testing.T) {
	parserImpl := NewParserImpl()

	tests := []struct {
		name     string
		calendar [][]string
		want     []MatchResults
		wantErr  bool
	}{
		{
			name: "Single Match Day with Giornata Marker",
			calendar: [][]string{
				{"Giornata 1", "", "", "", "", "", "", "", "", ""},
				{"TeamA", "P1", "G1", "TeamB", "2-1", "TeamC", "P2", "G2", "TeamD", "0-0"},
			},
			want: []MatchResults{
				{
					TeamResults: []TeamResult{
						{Team: "TeamA", Goals: 2, Points: 3},
						{Team: "TeamB", Goals: 1, Points: 0},
						{Team: "TeamC", Goals: 0, Points: 1},
						{Team: "TeamD", Goals: 0, Points: 1},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple Match Days",
			calendar: [][]string{
				{"Giornata 1", "", "", "", "", "Giornata 2", "", "", "", ""},
				{"TeamA", "1", "1", "TeamB", "1-0", "TeamA", "1", "1", "TeamC", "1-0"},
				{"TeamC", "2", "2", "TeamD", "2-2", "TeamB", "4", "4", "TeamD", "3-0"},
			},
			want: []MatchResults{
				{
					TeamResults: []TeamResult{
						{Team: "TeamA", Goals: 1, Points: 3},
						{Team: "TeamB", Goals: 0, Points: 0},
						{Team: "TeamC", Goals: 2, Points: 1},
						{Team: "TeamD", Goals: 2, Points: 1},
					},
				},
				{
					TeamResults: []TeamResult{
						{Team: "TeamA", Goals: 1, Points: 3},
						{Team: "TeamC", Goals: 0, Points: 0},
						{Team: "TeamB", Goals: 3, Points: 3},
						{Team: "TeamD", Goals: 0, Points: 0},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "Empty Calendar",
			calendar: [][]string{},
			want:     nil,
			wantErr:  false,
		},
		{
			name: "Calendar with only Giornata markers",
			calendar: [][]string{
				{"Giornata 1", "", "", "", "", "", "", "", "", ""},
				{"Giornata 2", "", "", "", "", "", "", "", "", ""},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Calendar with invalid rows (Giornata 2 not played yet)",
			calendar: [][]string{
				{"Giornata 1", "", "", "", "", "Giornata 2", "", "", "", ""},
				{"TeamA", "P1", "G1", "TeamB", "2-1", "TeamA", "P1", "G1", "TeamC", ""},
				{"TeamC", "P1", "G1", "TeamD", "2-1", "TeamB", "P1", "G1", "TeamD", ""},
			},
			want: []MatchResults{
				{
					TeamResults: []TeamResult{
						{Team: "TeamA", Goals: 2, Points: 3},
						{Team: "TeamB", Goals: 1, Points: 0},
						{Team: "TeamC", Goals: 2, Points: 3},
						{Team: "TeamD", Goals: 1, Points: 0},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserImpl.GetTeamResults(tt.calendar)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTeamResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTeamResults() got = %v, want %v", got, tt.want)
			}
		})
	}
}
