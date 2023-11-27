# Simple Duolingo Words Scraper
## An educational project

----------

## How to use

Run:
`go run duo.go`

## How it works?
This project uses [Colly](https://github.com/gocolly/colly) package to scrape the data. In the internal [Duo library](/duo/scraper.go) there exists a wrapper which makes using the colly collector easier.

Two scrapers run consequently. The first one visits the [Duolingo Fandom Wiki](https://duolingo.fandom.com/wiki/Turkish_Skill:Basics) and gathers all available units as links, the second one visits each unit link and scrapes all available words in the unit's lessons.

The scraper eventually puts everything in a json file.

A command line progress bar is also shown as a good example of concurrency in go.