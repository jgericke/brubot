package sources

import (
	"brubot/internal/helpers"
	"fmt"
	"log"
	"strconv"

	"github.com/gocolly/colly"
)

// Source represents a source data location for margin retrieval.
type Source struct {
	SourceRound        Round
	SourceURL          string
	SourceSearchFormat string
	SourceSearchText   string
}

// Round represents a round of fixtures.
type Round struct {
	RoundNumber   int
	RoundFixtures []Fixture
}

// Fixture represents match details within a round.
type Fixture struct {
	TeamnOne             string
	TeamTwo              string
	TeamOneWinPercentage string
	TeamTwoWinPercentage string
	PredictedMargin      int
}

// setRoundNumber retrieves and sets the current round.
func (r *Round) setRoundNumber() error {
	var err error
	if r.RoundNumber, err = helpers.FindRoundNumber(); err != nil {
		return err
	}
	return nil
}

// setSourceSearchText sets the text to search sourceURL body for, using RoundNumber as an identifier.
func (s *Source) setSourceSearchText() {
	s.SourceSearchText = fmt.Sprintf(s.SourceSearchFormat, s.SourceRound.RoundNumber)
}

// SourceRugbyVision predicted margins.
func SourceRugbyVision() (Source, error) {

	var err error
	source := Source{}
	source.SourceURL = "http://www.rugbyvision.com/updates/super-rugby-predictions"
	source.SourceSearchFormat = "Super Rugby Round %d Predictions"

	source.SourceRound.setRoundNumber()
	source.setSourceSearchText()

	// Instantiate default collector.
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
		colly.CacheDir("./cache"),
	)
	c.IgnoreRobotsTxt = false

	c.OnRequest(func(r *colly.Request) {
		log.Println("Retrieving from:", r.URL)
	})

	// Callback when collector finds the entry point to the DOM segment after matching criteria.
	c.OnHTML("table[data-title='"+source.SourceSearchText+"'] tbody", func(e *colly.HTMLElement) {

		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {

			if predictedMargin, marginErr := strconv.Atoi(el.ChildText("td:nth-child(5)")); marginErr != nil {
				err = marginErr
			} else {
				source.SourceRound.RoundFixtures = append(source.SourceRound.RoundFixtures, Fixture{
					TeamnOne:             el.ChildText("td:nth-child(1)"),
					TeamOneWinPercentage: el.ChildText("td:nth-child(2)"),
					TeamTwoWinPercentage: el.ChildText("td:nth-child(3)"),
					TeamTwo:              el.ChildText("td:nth-child(4)"),
					PredictedMargin:      predictedMargin,
				})
			}
		})
	})

	// Collector error handling.
	c.OnError(func(r *colly.Response, respError error) {
		err = fmt.Errorf("Error response %+v occurred retrieving from %s message: %s", r, r.Request.URL, respError)
	})
	c.Visit(source.SourceURL)

	if err != nil {
		return source, err
	}
	return source, nil
}
