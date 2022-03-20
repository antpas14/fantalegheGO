package main

import "fmt"
import "log"
import "net/http"
import "strconv"
import "github.com/tebeka/selenium"
import "encoding/json"
import "strings"
import "regexp"


type TeamResult struct {
    TeamName string
    TeamPoints int
}

type Rank struct {
    Team string `json:"team"`
    EvPoints float64 `json:"evPoints"`
    Points int `json:"points"`
}

type Response struct {
    Status string `json:"status"`
    Message string `json:"message"`
    Rank []Rank `json:"rank"`
}

var wd selenium.WebDriver
var urlPrefix string

// Get map from calendar
func getCalendarMap(url string) map[int][]TeamResult {
    wd.NewSession()
    if err := wd.Get(url); err != nil {
        panic(err)
    }

    // This is to close the accept cookie popup...
    button, err := wd.FindElement(selenium.ByCSSSelector, ".css-47sehv")
    if err == nil {
        button.Click()
    }

    calendar, err := wd.FindElement(selenium.ByCSSSelector, ".calendar")
    if err != nil {
        panic(err)
    }
    calTxt, err := calendar.Text()
    if err != nil {
        panic(err)
    }

    // Preprocess HTML content to put every match day on a row
    giornataRegex := regexp.MustCompile(`(GIORNATA){1}`)
    calTxt = strings.Replace(calTxt, "\n", "", 1)
    calTxt = strings.Replace(calTxt, ")\n", "", -1)

    var acc string
    for _, line := range strings.Split(strings.TrimSuffix(calTxt, "\n"), "\n") {

        if (giornataRegex.MatchString(line)) {
            acc = acc + "\n"
        } else {
            acc = acc + line  + "| " // This won't work if team names have pipe as separator, good enough for now
        }
    }

    mapp := make(map[int][]TeamResult)
    for i, matchDay := range strings.Split(strings.TrimSuffix(acc, "\n"), "\n") {
        matchDayArray := strings.Split(matchDay, "|")
        if (len(matchDayArray) > 1) {
            if (isValidMatchDay(matchDayArray)) {
                mapp[i] = make([]TeamResult, 0)
                var teamIdx = 0
                for teamIdx < len(matchDayArray) - 2 {
                    list := mapp[i]
                    var teamResult TeamResult
                    teamName := matchDayArray[teamIdx]
                    teamName = strings.TrimSpace(teamName)
                    teamResult.TeamName = teamName
                    var teamPointIdx int
                    teamPointIdx = teamIdx + 1
                    teamPoint,_ := strconv.Atoi(strings.TrimSpace(matchDayArray[teamPointIdx]))

                    teamResult.TeamPoints = teamPoint
                    list = append(list, teamResult)
                    mapp[i] = list
                    teamIdx = teamIdx + 3
                }
            }
        }
    }
    selenium.DeleteSession(urlPrefix, wd.SessionID())

    return mapp;
}

func isValidMatchDay(matchDay []string) bool {
    // Test just the third element of the day
    _, err := strconv.ParseFloat(strings.TrimPrefix(matchDay[2], " "), 64);

    if err != nil {
        return false
    }
    return true
}

func getRankingMap(url string) map[string]int {
    var mapp = make(map[string]int)

    if err := wd.Get(url); err != nil {
        panic(err)
    }

    // This is to close the accept cookie popup...
    button, err := wd.FindElement(selenium.ByCSSSelector, ".css-47sehv")
    if err == nil {
        button.Click()
    }

    ranking, err := wd.FindElement(selenium.ByCSSSelector, ".table")
    if err != nil {
        panic(err)
    }
    rnkTxt, err := ranking.Text()

    for _, teamRank := range strings.Split(strings.TrimSuffix(rnkTxt, "\n"), "\n") {
        teamRankArray := strings.Split(teamRank,  " ")
        if isValidRank(teamRankArray) {
            pointsString := teamRankArray[len(teamRankArray) - 2]
            points,_ := strconv.Atoi(pointsString)
            // 11 elements if team name is one word, 12 if is two words...
            teamNameArray := teamRankArray[1:len(teamRankArray) - 9]
            teamName := strings.TrimSuffix(strings.Join(teamNameArray," "), " ")
            mapp[teamName] = points
        }

    }
    return mapp
}

func isValidRank(teamRank []string) bool {
    return teamRank[len(teamRank)-1] != "totali"
}

func calculate(ranking map[string]int, results map[int][]TeamResult) Response {
    var response Response
    var mapp = make(map[string]float64)

    for teamName,_ := range ranking {
        mapp[teamName] = 0
    }
    var combinations = float64(len(ranking) - 1)

    for _, teamResults := range results {
        for i, t1 := range teamResults {
            var expectedPointForTeamForMatch float64
            var pointsForAllCombinations float64
            for j, t2 := range teamResults {
                if (i != j) {
                    pointsForAllCombinations = pointsForAllCombinations + calculatePoints(t1, t2)
                }
            }
            expectedPointForTeamForMatch = pointsForAllCombinations / combinations
            mapp[t1.TeamName] = mapp[t1.TeamName] + expectedPointForTeamForMatch
        }
    }
    listRank := make([]Rank, 0)
    for teamName, teamEVPoints := range mapp {
        var rank Rank
        rank.Team = teamName
        rank.EvPoints = teamEVPoints
        rank.Points = ranking[teamName]
        listRank = append(listRank, rank)
    }

    response.Status = "ok"
    response.Message = "ok"
    response.Rank = listRank
    return response
}

func calculatePoints (t1 TeamResult, t2 TeamResult) float64 {
    if (t1.TeamPoints > t2.TeamPoints) {
        return 3;
    } else if (t1.TeamPoints < t2.TeamPoints) {
        return 0;
    } else {
        return 1;
    }
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	resultsData := getCalendarMap("https://leghe.fantacalcio.it/fanta-pescio/calendario")
	rankingData := getRankingMap("https://leghe.fantacalcio.it/fanta-pescio/classifica")

    response := calculate(rankingData, resultsData)
	jResponse, err := json.Marshal(response)
    if err != nil {
        // handle error
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(jResponse)
}

func initWebDriver() selenium.WebDriver {
    const (
        port            = 4444
        host            = "localhost"
    )

    // Connect to the WebDriver instance running locally.
    caps := selenium.Capabilities{"browserName": "firefox"}
    selenium.SetDebug(false)
    urlPrefix := fmt.Sprintf("http://%s:%d/wd/hub", host, port)
    wd, err := selenium.NewRemote(caps, urlPrefix)
    if err != nil {
        panic(err)
    }

    wd.ResizeWindow("", 724, 144340)
    fmt.Printf("FINISH INIT\n")

    return wd;
}

func main() {
    wd = initWebDriver()
	http.HandleFunc("/parse", parseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
