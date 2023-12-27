package calculate

import (
	"github.com/antpas14/fantalegheEV-api"
    "fantalegheGO/internal/config"
    "fantalegheGO/internal/fetcher"
    "fantalegheGO/internal/parser"
    "log"
    "sync"
)
type Calculate interface {
	GetRanks(url string) []api.Rank
}

type CalculateImpl struct{
}

// Create fetcher and parser instances at the module level
var configInstance, _ = config.LoadConfig()
var parserInstance = &parser.ParserImpl{}
var fetcherInstance = &fetcher.FetcherImpl{configInstance.FetcherUrl}

// GetRanks retrieves a list of ranks (api.Rank)
func (c *CalculateImpl) GetRanks(leagueName string) []api.Rank {
	// Retrieve raw data using fetcher
	rankingsRaw, _ := fetcherInstance.Retrieve(configInstance.BaseUrl + leagueName + configInstance.RankingUrl)
	calendarRaw, _ := fetcherInstance.Retrieve(configInstance.BaseUrl + leagueName + configInstance.CalendarUrl)

	rankings, err := parserInstance.GetPoints(rankingsRaw)
	results, err := parserInstance.GetResults(calendarRaw)

	if err != nil {
		log.Fatal(err)
	}

	return calculate(rankings, results)
}

func calculate(rankings map[string]int, results *sync.Map) []api.Rank {
	var evPointsMap = make(map[string]float64)

	for teamName, _ := range rankings {
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
				break
			}
			var expectedPointForTeamForMatch float64
			var pointsForAllCombinations float64

			for j, teamResult2 := range teamResults {
				if i != j {
					pointsForAllCombinations = pointsForAllCombinations + calculatePoints(teamResult1, teamResult2)
				}
			}
			expectedPointForTeamForMatch = pointsForAllCombinations / combinations
			evPointsMap[teamResult1.Name] = evPointsMap[teamResult1.Name] + expectedPointForTeamForMatch
		}
		return true
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

func calculatePoints(t1 parser.TeamResult, t2 parser.TeamResult) float64 {
	if t1.Points > t2.Points {
		return 3
	} else if t1.Points < t2.Points {
		return 0
	} else {
		return 1
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
