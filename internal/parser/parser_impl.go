package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type ParserImpl struct{}

func NewParserImpl() *ParserImpl {
	return &ParserImpl{}
}

func (p *ParserImpl) GetTeamResults(calendar [][]string) ([]MatchResults, error) {
	var teamResults []TeamResult
	var results []MatchResults

	for _, calendarRow := range splitRows(calendar) {
		if len(calendarRow) > 0 && strings.Contains(calendarRow[0], "Giornata") {
			if len(teamResults) > 0 {
				results = append(results, MatchResults{TeamResults: teamResults})
			}
			teamResults = []TeamResult{}
			continue
		}
		teamResults = append(teamResults, getTeamResult(calendarRow)...)
	}

	if len(teamResults) > 0 {
		results = append(results, MatchResults{TeamResults: teamResults})
	}

	return results, nil
}

func getTeamResult(match []string) []TeamResult {
	if len(match) < 5 {
		return []TeamResult{}
	}

	goalsStr := strings.Split(match[4], "-")
	if len(goalsStr) != 2 {
		return []TeamResult{}
	}

	if len(goalsStr[0]) == 0 || len(goalsStr[1]) == 0 {
		return []TeamResult{}
	}

	goalA, errA := strconv.Atoi(goalsStr[0])
	goalB, errB := strconv.Atoi(goalsStr[1])

	if errA != nil || errB != nil {
		fmt.Printf("Error parsing goals: %v, %v\n", errA, errB)
		return []TeamResult{}
	}

	teamA := match[0]
	teamB := match[3]

	return []TeamResult{
		{Team: teamA, Goals: goalA, Points: calculateMatchPoints(goalA, goalB)},
		{Team: teamB, Goals: goalB, Points: calculateMatchPoints(goalB, goalA)},
	}
}

func calculateMatchPoints(ourGoals, theirGoals int) int {
	if ourGoals > theirGoals {
		return 3
	} else if ourGoals < theirGoals {
		return 0
	}
	return 1
}

func splitRows(rows [][]string) [][]string {
	var firstHalves [][]string
	var secondHalves [][]string
	var result [][]string

	for _, innerList := range rows {
		if len(innerList) != 10 {
			continue
		}
		midpoint := len(innerList) / 2
		firstHalf := innerList[:midpoint]
		secondHalf := innerList[midpoint:]

		firstHalves = append(firstHalves, firstHalf)
		secondHalves = append(secondHalves, secondHalf)
	}

	result = append(result, firstHalves...)
	result = append(result, secondHalves...)

	return result
}
