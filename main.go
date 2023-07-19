package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgent = []string{
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
}

var semaphore = make(chan struct{}, len(userAgent))
var baseURL = "https://www.theguardian.com"

func main() {
	worklist := make(chan []string)
	go func() {
		worklist <- []string{baseURL}
	}()

	seen := make(map[string]bool)

	for links := range worklist {
		for _, link := range links {
			if !seen[link] {
				seen[link] = true

				go func(link string) {
					foundLinks := crawl(link)
					if foundLinks != nil {
						worklist <- foundLinks
					}
				}(link)
			}
		}
	}
}

func crawl(targetURL string) []string {
	fmt.Println(targetURL)

	semaphore <- struct{}{}
	resp, err := getRequest(targetURL)
	if err != nil {
		log.Fatal("Failed to get a response:", err)
	}
	<-semaphore

	foundURLs := []string{}

	links := discoverLinks(resp)
	for _, link := range links {
		correctLink, ok := resolveRelativeLinks(link)
		if ok && correctLink != "" {
			foundURLs = append(foundURLs, correctLink)
		}
	}

	ParseHTML(resp)
	return foundURLs
}

func ParseHTML(response *http.Response) {
	// TODO
}

func getRequest(targetURL string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func discoverLinks(response *http.Response) []string {
	if response != nil {
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			log.Fatal("Failed to parse response")
		}

		foundURLs := []string{}
		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				res, _ := s.Attr("href")
				foundURLs = append(foundURLs, res)
			})
		}
		return foundURLs
	}
	return []string{}
}

func resolveRelativeLinks(link string) (string, bool) {
	resultLink := checkRelative(link)
	baseParse, _ := url.Parse(baseURL)
	resultParse, _ := url.Parse(resultLink)

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return resultLink, true
		}
	}
	return "", false
}

func checkRelative(link string) string {
	if strings.HasPrefix(link, "/") {
		return fmt.Sprintf("%s%s", baseURL, link)
	}
	return link
}

func randomUserAgent() string {
	seed := time.Now().Unix()
	r := rand.New(rand.NewSource(seed))
	randNum := r.Int() % len(userAgent)
	return userAgent[randNum]
}
