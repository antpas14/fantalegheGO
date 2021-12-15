package main

import "fmt"
import "log"
//import "io"
import "net/http"
//import "github.com/PuerkitoBio/goquery"
// import "strings"
import "strconv"

import 	"github.com/tebeka/selenium"
import "os"

type html struct {
	Body body `xml:"body"`
}

type body struct {
	Content string `xml:",innerxml"`
}

type teamResult struct {
    TeamName string
    TeamPoints int
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func getMap(url string) map[string][]teamResult {
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

    	// Navigate to the simple playground interface.
    	if err := wd.Get(url); err != nil {
    		panic(err)
    	}

    	// Get the list of matches
    	matches, err := wd.FindElements(selenium.ByCSSSelector, ".match")
    	if err != nil {
    		panic(err)
    	}

        mapp := make(map[string][]teamResult)
        // For each match, get the score
    	for _, match := range matches {
    	    teams, err := match.FindElements(selenium.ByCSSSelector, ".team")
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
                    fmt.Printf("%d %s", teamScore, teamName);
                }
            }
        }
        return mapp;
}

/* private String getTeamNameFromMatch(Element t) {
        return t.select(".team-name").get(0).text();
    } */

/*     private Integer getTeamPointsFromMatch(Element t) {
        return Integer.parseInt(t.select(".team-score").get(0).text());
    } */

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
	getMap("https://leghe.fantacalcio.it/fanta-pescio/calendario")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/parse", parseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
