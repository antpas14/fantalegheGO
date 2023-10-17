package calculate

import (
	"github.com/antpas14/fantalegheEV-api"

    "fantalegheGO/internal/config"
    "fantalegheGO/internal/fetcher"
    "fantalegheGO/internal/parser"

    "fmt"
    "log"
    "sync"

)

// GetRanks retrieves a list of ranks (api.Rank)
func GetRanks(leagueName string) []api.Rank {
    config, err := config.LoadConfig();
    if err != nil {
        log.Fatal(err)
    }

	parser := &parser.ParserImpl{}
    fetcher := &fetcher.FetcherImpl{config.FetcherUrl}
    // Retrieve raw data using fetcher
    rankingsRaw, _ := fetcher.Retrieve(config.BaseUrl + leagueName + config.RankingUrl)
    calendarRaw, _ := fetcher.Retrieve(config.BaseUrl + leagueName + config.CalendarUrl)


    rankings, err := parser.GetPoints(rankingsRaw)
    results, err := parser.GetResults(calendarRaw)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("LeagueName is %s", leagueName)

	// Perform data retrieval and processing here
	// For this example, we'll return a static list of ranks
	return calculate(rankings, results)
}

func calculate(rankings map[string]int, results *sync.Map) []api.Rank {
    var evPointsMap = make(map[string]float64)

    for teamName,_ := range rankings {
        evPointsMap[teamName] = 0
    }
    var combinations = float64(len(rankings) - 1)
    var teamResults []parser.TeamResult

    results.Range(func(_, value interface{}) bool {
        teamResults = make([]parser.TeamResult, 0)
        if trSlice, ok := value.([]parser.TeamResult); ok {
            teamResults = append(teamResults, trSlice...)
        }
        for i, teamResult1 := range teamResults {
            if teamResult1.Points == -1 {
                break;
            }
            var expectedPointForTeamForMatch float64
            var pointsForAllCombinations float64

            for j, teamResult2 := range teamResults {
                if (i != j) {
                    pointsForAllCombinations = pointsForAllCombinations + calculatePoints(teamResult1, teamResult2)
                }
            }
            expectedPointForTeamForMatch = pointsForAllCombinations / combinations
            evPointsMap[teamResult1.Name] = evPointsMap[teamResult1.Name] + expectedPointForTeamForMatch
        }
        return true;
    })
    listRank := make([]api.Rank, 0)
    for teamName, teamEVPoints := range evPointsMap {
        points := rankings[teamName]

        // Create new variables for teamName and evPoints inside the loop
        teamNameCopy := teamName
        evPointsCopy := teamEVPoints

        rank := api.Rank{
            Team:     &teamNameCopy,
            EvPoints: &evPointsCopy,
            Points:   &points,
        }
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