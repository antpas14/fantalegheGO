package calculate

import (
	"github.com/antpas14/fantalegheEV-api"
    "fantalegheGO/internal/fetcher"
    "fantalegheGO/internal/parser"
    "fmt"
    "log"

)

// GetRanks retrieves a list of ranks (api.Rank)
func GetRanks(leagueName string) []api.Rank {
    // Retrieve raw data using fetcher
    rankingsRaw, _ := fetcher.Retrieve("https://leghe.fantacalcio.it/fanta-pescio/classifica")
    calendarRaw, _ := fetcher.Retrieve("https://leghe.fantacalcio.it/fanta-pescio/calendario")

	parser := &parser.ParserImpl{}

    rankings, err := parser.GetPoints(rankingsRaw)
    results, err := parser.GetResults(calendarRaw)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("LeagueName is %s", leagueName)

    fmt.Printf("points are %s", rankings)
    fmt.Printf("results are %s", results)

	// Perform data retrieval and processing here
	// For this example, we'll return a static list of ranks
	return calculate(rankings, results)
}

func calculate(rankings map[string]int, results map[int][]parser.TeamResult) []api.Rank {
    var mapp = make(map[string]float64)

    for teamName,_ := range rankings {
        mapp[teamName] = 0
    }
    var combinations = float64(len(rankings) - 1)

    for _, teamResults := range results {
        for i, t1 := range teamResults {
            if t1.Points == -1 {
                break
                }
            var expectedPointForTeamForMatch float64
            var pointsForAllCombinations float64
            for j, t2 := range teamResults {
                if (i != j) {
                    pointsForAllCombinations = pointsForAllCombinations + calculatePoints(t1, t2)
                }
            }
            expectedPointForTeamForMatch = pointsForAllCombinations / combinations
            mapp[t1.Name] = mapp[t1.Name] + expectedPointForTeamForMatch
        }
    }
    listRank := make([]api.Rank, 0)
    for teamName, teamEVPoints := range mapp {
        points := rankings[teamName]
        var rank api.Rank
        rank.Team = &teamName
        rank.EvPoints = &teamEVPoints
        rank.Points = &points
        listRank = append(listRank, rank)
    }

    return listRank
}

func calculatePoints (t1 parser.TeamResult, t2 parser.TeamResult) float64 {
    if (t1.Points > t2.Points) {
        return 3;
    } else if (t1.Points < t2.Points) {
        return 0;
    } else {
        return 1;
    }
}

// Helper function to create a float64 pointer
func float64Ptr(f float64) *float64 {
	return &f
}

// Helper function to create an int pointer
func intPtr(i int) *int {
	return &i
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}