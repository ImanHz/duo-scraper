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

type SafeMap[T any] struct {
	mu  sync.Mutex
	Map map[string][]map[string]T
}

const HOMEPAGE = "https://duolingo.fandom.com/wiki/Turkish_Skill:Basics"

func main() {
	Start()
}

func Start() {

	u, _ := url.Parse(HOMEPAGE)
	base := u.Scheme + "://" + u.Host + "/wiki/"

	urls := make([]string, 0, 100)
	cfg := duo.CollyConfigs{
		OnError: func(r *colly.Response, err error) {
			log.Error().Msg(r.Request.URL.String())
		},
		OnHTML: map[string]func(e *colly.HTMLElement){
			"table.duonavbox .hlist": func(e *colly.HTMLElement) {
				e.ForEach("a,strong", func(i int, h *colly.HTMLElement) {
					omitSpace := strings.ReplaceAll(h.Text, " ", "_")
					// omitSpace = strings.ReplaceAll(omitSpace, "'", "\\'")
					if !strings.Contains(omitSpace, "-") {
						omitSpace = url.QueryEscape(omitSpace)
					}
					urls = append(urls, omitSpace)
				})
			},
		},
		Finally: func() {
			urlList := make(map[string]string, len(urls))

			words := SafeMap[string]{Map: map[string][]map[string]string{}}
			for i, u := range urls {
				urlList[base+"Turkish_Skill:"+u] = fmt.Sprintf("%02d_%s", i+1, u)
			}

			scrapeLessons(urlList, &words)
			// writeToCSV("words.csv", &words)
			// wordsMap := make(map[string]string)
			// words.Range(func(key, value any) bool {
			// wordsMap[key.(string)] = value.(string)

			if b, jerr := json.Marshal(words.Map); jerr == nil {
				os.WriteFile("file.json", b, 0600)
			} else {
				log.Error().Str("error in serialization", jerr.Error())
			}

			// return true
			// })
		},
		Timeout: 5000,
		Retry:   0,
		URLS:    []string{HOMEPAGE},
	}

	cfg.Scrape()

	// fmt.Println(base)
}

func scrapeLessons(lessons map[string]string, words *SafeMap[string]) {

	urls := make([]string, 0, len(lessons))

	for k := range lessons {
		urls = append(urls, k)
	}

	// words := sync.Map{}
	cfg := duo.CollyConfigs{
		OnRequest: func(r *colly.Request) {

			// log.Info().Msg("Scraping..." + r.URL.String())
		},
		OnInit: func() {
			log.Info().Msg("Now scraping each page...")
		},
		OnError: func(r *colly.Response, err error) {

			// log.Error().Msg(r.Request.URL.String())
		},
		OnHTML: map[string]func(e *colly.HTMLElement){

			"#mw-content-text": func(e *colly.HTMLElement) {

				listOfWords := []map[string]string{}

				e.ForEach("ul", func(i int, h *colly.HTMLElement) {
					e.ForEach("li", func(i int, h *colly.HTMLElement) {

						s := strings.Split(h.Text, "=")
						if len(s) == 2 {
							key := strings.TrimSpace(s[0])
							value := strings.TrimSpace(s[1])

							listOfWords = append(listOfWords, map[string]string{key: value})
							// words.mu.Lock()
							// words.Map = append(words.Map, map[string]string{key: value})
							// words.mu.Unlock()

						}
					})
				},
				)

				// if decodedURL, err := url.QueryUnescape(e.Request.URL.String()); err == nil {

				words.mu.Lock()
				if u, ok := lessons[e.Request.URL.String()]; ok {
					words.Map[u] = listOfWords
				} else {
					// fmt.Print("\010")
					log.Debug().Msg(e.Request.URL.String() + "  " + u)
					// fmt.Print("\033[A\033[K")

				}
				words.mu.Unlock()
				// } else {
				// 	log.Error().Msg(err.Error())
				// }
				// }
			}},
		Finally: func() {

		},
		Timeout:     5000,
		Retry:       3,
		Parallelism: 16,
		URLS:        urls,
	}

	cfg.Scrape()
	log.Debug().Msgf("we have %d urls and %d data", len(urls), len(words.Map))

	// for k := range lessons {
	// 	fmt.Println(k)
	// }
}

// func encodeCSV(columns []string, rows *sync.Map) ([]byte, error) {
// 	var buf bytes.Buffer
// 	w := csv.NewWriter(&buf)
// 	if err := w.Write(columns); err != nil {
// 		return nil, err
// 	}

// 	rows.Range(func(key, value any) bool {
// 		r := []string{key.(string), value.(string)}
// 		if err := w.Write(r); err != nil {
// 			log.Error().Err(err)
// 			return false
// 		}
// 		return true
// 	})

// 	if w.Flush(); w.Error() != nil {
// 		return nil, w.Error()
// 	}
// 	return buf.Bytes(), nil
// }

// func writeToCSV(filename string, words *sync.Map) {
// 	cols := []string{"Word", "Definition"}
// 	if b, err := encodeCSV(cols, words); err == nil {
// 		if ferr := os.WriteFile(filename, b, 0600); ferr == nil {
// 			log.Info().Msg("success!")

// 		} else {
// 			log.Error().Err(ferr)
// 		}
// 	} else {
// 		log.Error().Err(err)
// 	}
// }
