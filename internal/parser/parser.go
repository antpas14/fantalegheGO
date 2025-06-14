package parser

type MatchResults struct {
	TeamResults []TeamResult
}

type TeamResult struct {
	Team   string
	Goals  int
	Points int
}

type Parser interface {
	GetTeamResults(calendar [][]string) ([]MatchResults, error)
}
