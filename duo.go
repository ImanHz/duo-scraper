package main

import (
	"duo-scraper/duo"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog/log"
)

// since we are scraping in async. mode, we should take the thread-safety into consideration
type SafeMap[T any] struct {
	mu  sync.Mutex
	Map map[string][]map[string]T
}

// we get the list of the urls to scrape from this link
const HOMEPAGE = "https://duolingo.fandom.com/wiki/Turkish_Skill:Basics"

// the app entry point
func main() {
	Start()
}

func Start() {

	u, _ := url.Parse(HOMEPAGE)

	// this is the base url, created from the main url
	base := u.Scheme + "://" + u.Host + "/wiki/"

	// we gather the unitTitles here
	unitTitles := make([]string, 0, 100)

	// settings for the first scraper which gets the list of urls for the next scraper
	cfg := duo.CollyConfigs{
		OnError: func(r *colly.Response, err error) {
			log.Error().Msg(r.Request.URL.String())
		},
		OnHTML: map[string]func(e *colly.HTMLElement){
			"table.duonavbox .hlist": func(e *colly.HTMLElement) {
				// we have N urls, one of which is a bold tag and the other N-1 are hrefs
				e.ForEach("a,strong", func(i int, h *colly.HTMLElement) {
					// we should replace the blank spaces with an underscore
					omitSpace := strings.ReplaceAll(h.Text, " ", "_")

					// some links contain hyphen, we should not escape them
					if !strings.Contains(omitSpace, "-") {
						omitSpace = url.QueryEscape(omitSpace)
					}
					unitTitles = append(unitTitles, omitSpace)
				})
			},
		},
		// after the first scraper is done, we call this
		Finally: func() {

			// the list of words will be added here, this is a global but safe map
			// structure:
			// {
			//   LESSON1 : {[WORD1:DEFINITION, WORD2:DEFINITION,]}
			//   LESSON2 : {[WORD1:DEFINITION, WORD2:DEFINITION,]}
			// }
			words := SafeMap[string]{Map: map[string][]map[string]string{}}

			unitsMap := make(map[string]string, len(unitTitles))

			// a map for keeping the lesson title and its url
			// KEY: URL , VALUE: the LESSON TITLE
			// everything is made based on the lesson titles which is scraped before
			for i, l := range unitTitles {
				unitsMap[base+"Turkish_Skill:"+l] = fmt.Sprintf("%02d_%s", i+1, l)
			}

			// pass the url map and address of the words map to scrape
			scrapeLessons(unitsMap, &words)

			// write everything to a json file
			if b, jerr := json.Marshal(words.Map); jerr == nil {
				os.WriteFile("file.json", b, 0600)
			} else {
				log.Error().Str("error in serialization", jerr.Error())
			}
		},
		Timeout: 5000,
		Retry:   0,
		URLS:    []string{HOMEPAGE},
	}

	cfg.Scrape()

}

func scrapeLessons(units map[string]string, words *SafeMap[string]) {

	// gathering the urls
	urls := make([]string, 0, len(units))

	for k := range units {
		urls = append(urls, k)
	}

	// setting up another scraper
	cfg := duo.CollyConfigs{

		OnInit: func() {
			log.Info().Msg("Now scraping each page...")
		},

		OnHTML: map[string]func(e *colly.HTMLElement){

			// these pages have a flat hierarchy which is hard to scrape
			// we use a trick here, find all lists, i.e. (ul, li)s and found the terms
			// which have =
			"#mw-content-text": func(e *colly.HTMLElement) {

				// first gather the words, then add it to the main map
				listOfWords := []map[string]string{}

				e.ForEach("ul", func(i int, h *colly.HTMLElement) {
					e.ForEach("li", func(i int, h *colly.HTMLElement) {

						s := strings.Split(h.Text, "=")
						// if the string is WORD=DEFINITION
						if len(s) == 2 {

							// cleaning the data
							key := strings.TrimSpace(s[0])
							value := strings.TrimSpace(s[1])

							listOfWords = append(listOfWords, map[string]string{key: value})

						}
					})
				},
				)

				words.mu.Lock()
				// find the lesson title from the input map based on the request URL
				if u, ok := units[e.Request.URL.String()]; ok {
					words.Map[u] = listOfWords
				} else {
					log.Debug().Msg(e.Request.URL.String() + "  " + u)
				}
				words.mu.Unlock()

			}},
		Timeout:     5000,
		Retry:       3,
		Parallelism: 16,
		URLS:        urls,
	}

	cfg.Scrape()
	log.Debug().Msgf("we have %d urls and %d data", len(urls), len(words.Map))

}
