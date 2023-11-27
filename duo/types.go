package duo

import (
	"github.com/gocolly/colly"
)

// this the configuration type to configure the colly scraper
type CollyConfigs struct {
	OnRequest   func(r *colly.Request)
	OnResponse  func(r *colly.Response)
	OnError     func(r *colly.Response, err error)
	OnHTML      map[string]func(e *colly.HTMLElement) // key: the query string, value: the function to call when query was found
	OnInit      func()                                // being called before running the colly
	Finally     func()                                // being called when everything is finished
	Timeout     int                                   // in milliseconds
	Parallelism int                                   // 0 for sync. scraping
	Retry       int
	URLS        []string
	DomainGlob  string // defauly is "*"
}

// a simple error wrapper
type Error struct {
	E     error
	Where string
}
