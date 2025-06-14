package calculate

import (
	"fantalegheGO/internal/excel"
	"fantalegheGO/internal/parser"
	"fmt"
	"mime/multipart"
	"sort"

	api "github.com/antpas14/fantalegheEV-api"
)

type CalculateImpl struct {
	excelService excel.ExcelService // Changed to interface
	parser       parser.Parser      // Changed to interface
}

type evRankData struct {
	EvSum       float64
	TotalPoints int
	MatchCount  int
}

// NewCalculateImpl now takes interfaces
func NewCalculateImpl(es excel.ExcelService, p parser.Parser) *CalculateImpl {
	return &CalculateImpl{
		excelService: es,
		parser:       p,
	}
}

func (c *CalculateImpl) GetRanks(fileHeader *multipart.FileHeader) ([]api.Rank, error) {
	excelRawData, err := c.excelService.ReadExcel(fileHeader)
	if err != nil {
		// Wrap the error to provide more context.
		return nil, fmt.Errorf("failed to read excel file: %w", err)
	}

	results, err := c.parser.GetTeamResults(excelRawData)
	if err != nil {
		return nil, fmt.Errorf("failed to get team results: %w", err)
	}

	finalRanks := calculate(results)
	return finalRanks, nil
}

func calculate(results []parser.MatchResults) []api.Rank {
	evRankMap := make(map[string]evRankData)

	for _, matchResult := range results {
		for i := 0; i < len(matchResult.TeamResults); i++ {
			t1 := matchResult.TeamResults[i]
			currentMatchDayPoints := float64(0)

			for j := 0; j < len(matchResult.TeamResults); j++ {
				if i != j {
					t2 := matchResult.TeamResults[j]
					currentMatchDayPoints += calculatePoints(t1, t2)
				}
			}
			currentEvData, ok := evRankMap[t1.Team]
			if !ok {
				currentEvData = evRankData{EvSum: 0.0, TotalPoints: 0, MatchCount: 0}
			}

			if len(matchResult.TeamResults) > 0 {
				currentEvData.EvSum += currentMatchDayPoints / float64(len(matchResult.TeamResults)-1)
			}
			currentEvData.TotalPoints += t1.Points
			currentEvData.MatchCount++

			evRankMap[t1.Team] = currentEvData
		}
	}

	var ranks []api.Rank
	for teamName, data := range evRankMap {
		ranks = append(ranks, api.Rank{
			Team:     &teamName,
			EvPoints: &data.EvSum,
			Points:   &data.TotalPoints,
		})
	}

	sort.Slice(ranks, func(i, j int) bool {
		return *ranks[i].EvPoints > *ranks[j].EvPoints
	})

	return ranks
}

func calculatePoints(t1 parser.TeamResult, t2 parser.TeamResult) float64 {
	if t1.Goals > t2.Goals {
		return 3
	} else if t1.Goals < t2.Goals {
		return 0
	} else {
		return 1
	}
}
