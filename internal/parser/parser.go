package parser

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
// 	"fmt"
)

type Parser interface {
	GetPoints(rankingsPage string) (map[string]int, error)
	GetResults(calendarPage string) (map[int][]TeamResult, error)
}

type TeamResult struct {
	Name   string
	Points int
}

type ParserImpl struct{}

func (p *ParserImpl) GetPoints(rankingsPage string) (map[string]int, error) {
    doc, err := goquery.NewDocumentFromReader(strings.NewReader(rankingsPage))
	if err != nil {
		return nil, err
	}

	rankingTable := p.getRankingTable(doc)

	pointsMap := make(map[string]int)
//     fmt.Printf("Ranking page %s", rankingsPage)
//     fmt.Printf("Ranking table %s", rankingTable)
	rankingTable.Each(func(i int, s *goquery.Selection) {
		teamName := p.getTeamNameFromRankingTable(s)
		teamPoints := p.getTeamPointsFromRankingTable(s)
		pointsMap[teamName] = teamPoints
	})

	return pointsMap, nil
}

func (p *ParserImpl) GetResults(calendarPage string) (map[int][]TeamResult, error) {
doc, err := goquery.NewDocumentFromReader(strings.NewReader(calendarPage))
	if err != nil {
		return nil, err
	}

	calendarDays := p.selectCalendarDaysFromCalendarDocument(doc)
	resultsMap := make(map[int][]TeamResult)
	var counter int32

	var wg sync.WaitGroup

	calendarDays.Each(func(i int, s *goquery.Selection) {
		wg.Add(1)
		go func(index int, selection *goquery.Selection) {
			defer wg.Done()
			matches := p.getMatchesFromCalendarDay(selection)
			var teamResults []TeamResult

			matches.Each(func(j int, match *goquery.Selection) {
				teams := p.getTeamsFromMatches(match)
				teams.Each(func(k int, team *goquery.Selection) {
					teamName := p.getTeamNameFromMatch(team)
					teamPoints := p.getTeamPointsFromMatch(team)
					teamResults = append(teamResults, TeamResult{Name: teamName, Points: teamPoints})
				})
			})

			atomic.AddInt32(&counter, 1)
			resultsMap[int(atomic.LoadInt32(&counter))] = teamResults
		}(i, s)
	})

	wg.Wait()

	return resultsMap, nil
}

// Helper functions

func (p *ParserImpl) selectCalendarDaysFromCalendarDocument(doc *goquery.Document) *goquery.Selection {
	return doc.Find(".match-frame")
}

func (p *ParserImpl) getMatchesFromCalendarDay(calendarDay *goquery.Selection) *goquery.Selection {
	return calendarDay.Find(".match")
}

func (p *ParserImpl) getTeamsFromMatches(matches *goquery.Selection) *goquery.Selection {
    return matches.Find(".team")
}

func (p *ParserImpl) getTeamNameFromMatch(team *goquery.Selection) string {
    return team.Find(".team-name").First().Text()
}

func (p *ParserImpl) getTeamPointsFromMatch(team *goquery.Selection) int {
	teamFPT, _ := team.Find(".team-fpt").First().Html()
    	if val, err := strconv.ParseFloat(teamFPT, 64); err == nil && val > 0.0 {
    		teamScore, _ := team.Find(".team-score").First().Html()
    		points, _ := strconv.Atoi(teamScore)
    		return points
    	}
    	return -1
}

func (p *ParserImpl) getTeamNameFromRankingTable(e *goquery.Selection) string {
    return e.Children().Eq(2).Children().Eq(0).Children().Eq(0).Text()

}

func (p *ParserImpl) getTeamPointsFromRankingTable(e *goquery.Selection) int {
	teamPoints, _ := e.Children().Eq(10).Children().Eq(0).Html()
    	points, _ := strconv.Atoi(teamPoints)
    	return points
}

func (p *ParserImpl) getRankingTable(doc *goquery.Document) *goquery.Selection {
	return doc.Find(".ranking").Eq(0).Children().Eq(0).Children().Eq(1).Children()
}