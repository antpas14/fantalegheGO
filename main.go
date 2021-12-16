package main

import "fmt"
import "log"
import "net/http"
import "strconv"

import 	"github.com/tebeka/selenium"
import "os"
import "encoding/json"

type TeamResult struct {
    TeamName string
    TeamPoints int
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func getMap(url string) map[int][]TeamResult {
	// Request the HTML page.
	// Start a Selenium WebDriver server instance (if one is not already
    	// running).
    	const (
    		// These paths will be different on your system.
    		seleniumPath    = "/home/antonello/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/selenium-server.jar"
    		geckoDriverPath = "/home/antonello/go/pkg/mod/github.com/tebeka/selenium@v0.9.9/vendor/geckodriver"
    		port            = 4444
    	)
    	opts := []selenium.ServiceOption{
    		selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
    		selenium.GeckoDriver(geckoDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
    		selenium.Output(os.Stderr),            // Output debug information to STDERR.
    	}
    	selenium.SetDebug(false)
    	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
    	if err != nil {
    		panic(err) // panic is used only as an example and is not otherwise recommended.
    	}
    	defer service.Stop()

    	// Connect to the WebDriver instance running locally.
    	caps := selenium.Capabilities{"browserName": "firefox"}
    	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
    	if err != nil {
    		panic(err)
    	}
    	defer wd.Quit()
        wd.ResizeWindow("", 724, 144340)

    	if err := wd.Get(url); err != nil {
    		panic(err)
    	}

    	// This is to close the accept cookie popup...
    	button, err := wd.FindElement(selenium.ByCSSSelector, ".css-47sehv")
        if err == nil {
            // panic(err)
            button.Click()
        }


    	calendar, err := wd.FindElement(selenium.ByCSSSelector, ".calendar")
    	if err != nil {
    		panic(err)
    	}

    	// Get the list of matches
    	matches, err := calendar.FindElements(selenium.ByCSSSelector, ".match-results")
    	if err != nil {
    		panic(err)
    	}

        mapp := make(map[int][]TeamResult)
        // For each match, get the score
    	for i, match := range matches {
    	    teams, err := match.FindElements(selenium.ByCSSSelector, ".team")
            if err != nil {
                panic(err)
            }
            if err != nil {
                panic(err)
            }
            for _, team := range teams {
                teamName := getTeamNameFromMatch(team)
                teamScoreString := getTeamScoreFromMatch(team)
                if isValidResult(teamName, teamScoreString) {
                    teamScore, err := strconv.Atoi(teamScoreString);
                    if err != nil {
                        panic(err)
                    }
                    if _, ok := mapp[i]; !ok {
                        mapp[i] = make([]TeamResult, 0)
                    }

                    list := mapp[i]
                    var teamResult TeamResult
                    teamResult.TeamName = teamName
                    teamResult.TeamPoints = teamScore
                    list = append(list, teamResult)
                    mapp[i] = list
                } else {
                    break
                }
            }
        }
        return mapp;
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
	data := getMap("https://leghe.fantacalcio.it/fanta-pescio/calendario")

	jData, err := json.Marshal(data)
    if err != nil {
        // handle error
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(jData)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/parse", parseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
