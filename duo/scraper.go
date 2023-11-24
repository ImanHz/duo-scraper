package duo

import (
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const E_GREY = "\u26AA"
const E_GREEN = "\U0001F7E2"
const E_RED = "\U0001F534"

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("DuoLingo Scraper")
}

func (cfg *CollyConfigs) Scrape() {

	c := colly.NewCollector()

	if cfg.Parallelism > 0 {
		c.Async = true
		c.Limit(&colly.LimitRule{Parallelism: cfg.Parallelism})
	}

	if cfg.DomainGlob != "" {
		c.Limit(&colly.LimitRule{DomainGlob: cfg.DomainGlob})
	} else {
		c.Limit(&colly.LimitRule{DomainGlob: "*"})
	}

	if cfg.OnInit != nil {
		cfg.OnInit()
	}

	for range cfg.URLS {
		fmt.Print(E_GREY)

	}
	fmt.Print("\r")

	progress := make(chan Error, len(cfg.URLS))

	c.SetRequestTimeout(time.Millisecond * time.Duration(cfg.Timeout))
	for tag, f := range cfg.OnHTML {
		c.OnHTML(tag, f)
	}

	c.OnRequest(func(r *colly.Request) {

		if cfg.Retry > 0 {
			id := r.URL.String()
			if r.Ctx.GetAny(id) == nil {
				r.Ctx.Put(id, 0)
			}
		}

		if cfg.OnRequest != nil {
			cfg.OnRequest(r)
		}
	})

	c.OnResponse(func(r *colly.Response) {

		progress <- Error{E: nil, Where: r.Request.URL.String()}
		if cfg.OnResponse != nil {
			cfg.OnResponse(r)
		}
	})

	c.OnError(func(r *colly.Response, err error) {

		if cfg.Retry > 0 {
			id := r.Request.URL.String()
			ret := r.Ctx.GetAny(id)
			if ret != nil {

				retry := ret.(int)
				if retry < cfg.Retry {
					r.Request.Retry()
				} else {
					// log.Error().Str("error loading page", err.Error())
					progress <- Error{E: err, Where: r.Request.URL.String()}

				}
				retry++
				r.Ctx.Put(id, retry)

			}
		}
		if cfg.OnError != nil {
			cfg.OnError(r, err)
		}
	})

	for _, url := range cfg.URLS {
		if err := c.Visit(url); err != nil {
			log.Error().Msg(fmt.Sprintf("error visiting %s, error: %v"+url, err))
		}
	}
	errors := make([]Error, 0, len(cfg.URLS))
	for range cfg.URLS {
		state := <-progress
		if state.E == nil {
			fmt.Print(E_GREEN)
		} else {
			fmt.Print(E_RED)
			errors = append(errors, state)
		}
	}
	c.Wait()
	fmt.Println()

	for _, e := range errors {
		log.Error().Msg(fmt.Sprintf("Error in %s, error: %v", e.Where, e.E))
	}
	if cfg.Finally != nil {
		cfg.Finally()
	}

}
