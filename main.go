package main

import "fmt"
import "log"
import "net/http"
import "strconv"

import 	"github.com/tebeka/selenium"
import "encoding/json"
import "strings"
import "regexp"


type TeamResult struct {
    TeamName string
    TeamPoints int
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
    calTxt = strings.Replace(calTxt, ")\n", "", -1)
    if err != nil {
        panic(err)
    }

    giornataRegex := regexp.MustCompile(`(GIORNATA){1}`)
    var acc string
    for _, line := range strings.Split(strings.TrimSuffix(calTxt, "\n"), "\n") {

        if (giornataRegex.MatchString(line)) {
            acc = acc + "\n"
        } else {
            acc = acc + line  + "| " // This won't work if team names have pipe as separator, good enough for now
        }
    }
    acc = strings.Replace(acc, "\n", "", 1)

    mapp := make(map[int][]TeamResult)
    for i, matchDay := range strings.Split(strings.TrimSuffix(acc, "\n"), "\n") {
        matchDayArray := strings.Split(matchDay, "|")
        if (len(matchDayArray) > 1) {
            if (isValidMatchDay(matchDayArray)) {
                mapp[i] = make([]TeamResult, 0)
                var team = 0
                for team < len(matchDayArray) - 2 {
                    list := mapp[i]
                    var teamResult TeamResult
                    teamResult.TeamName = matchDayArray[team]
                    var teamPointIdx int
                    teamPointIdx = team + 1
                    teamPoint,_ := strconv.Atoi(strings.TrimPrefix(matchDayArray[teamPointIdx], " "))

                    teamResult.TeamPoints = teamPoint
                    list = append(list, teamResult)
                    mapp[i] = list
                    team = team + 3
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

func getTeamScoreFromMatch(element selenium.WebElement) string {
    score, err := element.FindElement(selenium.ByCSSSelector, ".team-score")
    if err != nil {
        log.Fatal(err)
    }
    teamScore, err := score.Text()
    if err != nil {
        log.Fatal(err)
    }
    return teamScore;
}

func getTeamNameFromMatch(element selenium.WebElement) string {
    team, err := element.FindElement(selenium.ByCSSSelector, ".team-name")
    if err != nil {
        log.Fatal(err)
    }
    teamName, err := team.Text()
    if err != nil {
        log.Fatal(err)
    }
    return teamName
}

func isValidResult(teamName string, teamScore string) bool {
    return len(teamName) > 0 && len(teamScore) > 0
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	data := getCalendarMap("https://leghe.fantacalcio.it/fanta-pescio/calendario")

	jData, err := json.Marshal(data)
    if err != nil {
        // handle error
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(jData)
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
