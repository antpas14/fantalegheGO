package main

import (
	"bytes"
	"encoding/xml"
	"fmt"

)

import "log"
import "net/http"
import "io/ioutil"
import "github.com/PuerkitoBio/goquery"

type html struct {
	Body body `xml:"body"`
}

type body struct {
	Content string `xml:",innerxml"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func getPage(url string)  body {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func parsePage(content string) body {
	b := []byte(content)
	h := html{}
	err := xml.NewDecoder(bytes.NewBuffer(b)).Decode(&h)
	if err != nil {
		fmt.Println("error", err)
		return h.Body
	}
	return h.Body
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
	page := getPage("https://leghe.fantacalcio.it/fanta-pescio/calendario")
    doc, err := goquery.NewDocumentFromReader(page)

    doc.Find(".match").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		fmt.Printf("Match %s\n", s)
	})
	fmt.Fprintf(w, parsePage(page))

}
func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, getPage("https://leghe.fantacalcio.it/fanta-pescio/calendario"))
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/test", test)
	http.HandleFunc("/parse", test)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
