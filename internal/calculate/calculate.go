package calculate

import (
	api "github.com/antpas14/fantalegheEV-api"

	"fantalegheGO/internal/config"
	"fantalegheGO/internal/fetcher"
	"fantalegheGO/internal/parser"
	"net/http"
	"sync"
)

type Calculate interface {
	GetRanks(url string) []api.Rank
}

type CalculateImpl struct {
}

// Create fetcher and config instances at the module level
var configInstance, _ = config.LoadConfig()
var fetcherInstance = &fetcher.FetcherImpl{configInstance.FetcherUrl, http.DefaultClient}

// GetRanks retrieves a list of ranks (api.Rank)
func (c *CalculateImpl) GetRanks(leagueName string, parserInstance parser.Parser) ([]api.Rank, error) {
	// Retrieve raw data using fetcher
	rankingsRaw, _ := fetcherInstance.Retrieve(configInstance.BaseUrl + leagueName + configInstance.RankingUrl)
	calendarRaw, _ := fetcherInstance.Retrieve(configInstance.BaseUrl + leagueName + configInstance.CalendarUrl)

	rankings, err := parserInstance.GetPoints(rankingsRaw)
	if err != nil {
		return nil, err
	}

	results, err := parserInstance.GetResults(calendarRaw)
	if err != nil {
		return nil, err
	}

	return calculate(rankings, results), nil
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
