package duo

import (
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// some emojis to show the scraping progress
const E_BLACK = "\u2B1B"
const E_GREEN = "\U0001F7E9"
const E_RED = "\U0001F7E5"

// initializes the logger
func init() {
	// changing the default json output to a more human-readable format
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Info().Msg("DuoLingo Scraper")
}

// this is wrapper for a typical scraper with retry, timeout and parallelism
// you should provide a proper CollyConfigs instance and call the Scrape()
func (cfg *CollyConfigs) Scrape() {

	// enable async. collecting by default
	c := colly.NewCollector(colly.Async(true))

	c.Limit(&colly.LimitRule{Parallelism: cfg.Parallelism})

	if cfg.DomainGlob != "" {
		c.Limit(&colly.LimitRule{DomainGlob: cfg.DomainGlob})
	} else {
		c.Limit(&colly.LimitRule{DomainGlob: "*"})
	}

	// runs this function before everything else
	if cfg.OnInit != nil {
		cfg.OnInit()
	}

	c.SetRequestTimeout(time.Millisecond * time.Duration(cfg.Timeout))

	// progress bar init.
	for range cfg.URLS {
		fmt.Print(E_BLACK)
	}
	fmt.Print("\r")

	// here we keep the progress by adding the OK or ERROR status
	progress := make(chan Error, len(cfg.URLS))

	// assigning the OnHTML callbacks and their corresponding quert strings
	for tag, f := range cfg.OnHTML {
		c.OnHTML(tag, f)
	}

	c.OnRequest(func(r *colly.Request) {

		// on each request, we add the key-value pair to the colly context
		// to keep track of the retries has made
		// here we initialize the value to 0
		if cfg.Retry > 0 {
			id := r.URL.String()
			if r.Ctx.GetAny(id) == nil {
				r.Ctx.Put(id, 0)
			}
		}

		// null-check before calling the function
		if cfg.OnRequest != nil {
			cfg.OnRequest(r)
		}
	})

	c.OnResponse(func(r *colly.Response) {

		// receiving the response means that we don't have any errors
		// note: OnHTML is being called after OnResponse!
		progress <- Error{E: nil, Where: r.Request.URL.String()}

		if cfg.OnResponse != nil {
			cfg.OnResponse(r)
		}
	})

	c.OnError(func(r *colly.Response, err error) {

		// on error, check the retry count, retry if the retry limit is not reached yet,
		// adds an error if the limit is reached
		if cfg.Retry > 0 {
			// getting the current retry count from the context
			id := r.Request.URL.String()
			ret := r.Ctx.GetAny(id)
			if ret != nil {
				retry := ret.(int)
				if retry < cfg.Retry {
					r.Request.Retry()
				} else {
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

	// visiting the urls
	for _, url := range cfg.URLS {
		if err := c.Visit(url); err != nil {
			log.Error().Msg(fmt.Sprintf("error visiting %s, error: %v"+url, err))
		}
	}

	// a slice to gather the values inside the progress channel
	errors := make([]Error, 0, len(cfg.URLS))

	// waiting for the colly to gather everything and showing the progress
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

	// log the existing errors
	for _, e := range errors {
		log.Error().Msg(fmt.Sprintf("Error in %s, error: %v", e.Where, e.E))
	}

	// running the finally callback
	if cfg.Finally != nil {
		cfg.Finally()
	}

}
