package parser

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// DocumentFromReaderProvider is an interface that abstracts goquery.NewDocumentFromReader function.
type DocumentFromReaderProvider interface {
	NewDocumentFromReader(r *strings.Reader) (*goquery.Document, error)
}

// RealDocumentProvider implements DocumentFromReaderProvider using the actual goquery implementation.
type DocumentProviderImpl struct{}

func (r *DocumentProviderImpl) NewDocumentFromReader(reader *strings.Reader) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(reader)
}

func DefaultParserImpl() *ParserImpl {
	return &ParserImpl{
		DocumentProvider: &DocumentProviderImpl{},
	}
}

func CustomParserImpl(documentProvider DocumentFromReaderProvider) *ParserImpl {
	return &ParserImpl{
		DocumentProvider: documentProvider,
	}
}

type Parser interface {
	GetPoints(rankingsPage string) (map[string]int, error)
	GetResults(calendarPage string) (*sync.Map, error)
}

type TeamResult struct {
	Name   string
	Points int
}

type ParserImpl struct {
	DocumentProvider DocumentFromReaderProvider
}

func (p *ParserImpl) GetPoints(rankingsPage string) (map[string]int, error) {
	doc, err := p.DocumentProvider.NewDocumentFromReader(strings.NewReader(rankingsPage))
	if err != nil {
		return nil, err
	}

	rankingTable := p.getRankingTable(doc)

	pointsMap := make(map[string]int)
	rankingTable.Each(func(i int, s *goquery.Selection) {
		teamName := p.getTeamNameFromRankingTable(s)
		teamPoints := p.getTeamPointsFromRankingTable(s)
		pointsMap[teamName] = teamPoints
	})

	return pointsMap, nil
}

func (p *ParserImpl) GetResults(calendarPage string) (*sync.Map, error) {
	doc, err := p.DocumentProvider.NewDocumentFromReader(strings.NewReader(calendarPage))
	resultsMap := &sync.Map{}

	if err != nil {
		return resultsMap, err
	}

	calendarDays := p.selectCalendarDaysFromCalendarDocument(doc)
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
					if teamPoints != -1 {
						teamResults = append(teamResults, TeamResult{Name: teamName, Points: teamPoints})
					}
				})
			})

			if len(teamResults) > 0 {
				atomic.AddInt32(&counter, 1)
				strAtomic := strconv.FormatInt(int64(atomic.LoadInt32(&counter)), 10)
				resultsMap.Store(strAtomic, teamResults)
			}
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
	var teamPoints = -1
	if val, err := strconv.ParseFloat(teamFPT, 64); err == nil && val > 0.0 {
		teamScore, _ := team.Find(".team-score").First().Html()
		points, _ := strconv.Atoi(teamScore)
		teamPoints = points
	}
	return teamPoints
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
