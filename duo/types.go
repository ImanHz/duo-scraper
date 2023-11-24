package duo

import (
	"github.com/gocolly/colly"
)

type CollyConfigs struct {
	OnRequest   func(r *colly.Request)
	OnResponse  func(r *colly.Response)
	OnError     func(r *colly.Response, err error)
	OnHTML      map[string]func(e *colly.HTMLElement)
	OnInit      func()
	Finally     func()
	Timeout     int
	Retry       int
	Parallelism int
	URLS        []string
	DomainGlob  string
}

type Error struct {
	E     error
	Where string
}
