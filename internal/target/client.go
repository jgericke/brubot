package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// Colly client for interacting with target
type client struct {
	collector *colly.Collector // colly client
	config    clientConfig     // colly client settings (http/TLS timeouts)
	parser    clientParser     // identifies fields to be scraped and parsed from target endpoints
}

// Client configuration
type clientConfig struct {
	urls                map[string]string // holds all url endpoints for a specific target (fixtures and predictions)
	userAgent           string            // set globally or per client
	ignoreRobots        bool              // colly parameter (true, who cares about robots anyway?)
	enableCache         bool              // colly parameter (true, caching is good)
	cacheDir            string            // colly parameter (directory path for cache storage)
	dialTimeout         time.Duration     // colly parameter (request timeout seconds)
	tlsHandShakeTimeout time.Duration     // colly parameter (tls handshake timeout seconds)
}

// Attributes and strings to be scraped and parsed
// on various target endpoints
type clientParser struct {
	login       map[string]string // string identifiers for target login page
	fixtures    map[string]string // string identifiers for target fixture attributes
	predictions map[string]string // string identifiers for target prediction query arguments
}

// Initialise colly client with clientConfig parameters
func (c *client) init(cookieJar http.CookieJar) error {

	// Creates and configures a colly instance with caching
	if c.config.enableCache {
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
			colly.CacheDir(c.config.cacheDir),
		)
	} else {
		// Creates and configures a colly instance without caching
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
		)
	}
	// Sets transport and TLS timeouts,
	// these may need to be relaxed in brubots config.yaml
	// if frequent timeouts occur.
	c.collector.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Second * c.config.dialTimeout,
		}).DialContext,
		TLSHandshakeTimeout: time.Second * c.config.tlsHandShakeTimeout,
	})

	c.collector.IgnoreRobotsTxt = c.config.ignoreRobots

	// Authentication to target is handled within auth.go
	// confirm at a minimum the cookie jar housing authentication
	// token is not empty before setting (we have missed auth in that case).
	// Note, its still possible the cookie has expired and if so, retry auth.
	if cookieJar != nil {
		c.collector.SetCookieJar(cookieJar)
	} else {
		return errors.New("An error occurred setting cookieJar on client, cookieJar is empty")
	}

	return nil

}

// getFixtures uses a pre-authenticated client to extract fixture parameters for a specified round.
func (c *client) getFixtures(round *Round) error {

	var err error

	// Scrapes and parses fixtures for the active round (set via round.id).
	// Fixture parameters are stored using the Endpoint.Fixture structure:
	//
	// 		token     string
	// 		leftTeam  string
	// 		rightTeam string
	// 		leftID    int
	// 		rightID   int
	// 		winnerID  int
	// 		margin    int
	//
	c.collector.OnHTML(c.parser.fixtures["attr_onhtml"], func(e *colly.HTMLElement) {

		e.ForEach(c.parser.fixtures["attr_fixture"], func(_ int, cl *colly.HTMLElement) {

			// Convert teamIDs to int (for better living).
			leftID, convErr := strconv.Atoi(cl.Attr(c.parser.fixtures["attr_t_leftid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			rightID, convErr := strconv.Atoi(cl.Attr(c.parser.fixtures["attr_t_rightid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			// Appends a fixture element to a slice of Fixtures
			// within the active Round, setting scraped and parsed fixture
			// parameters.
			round.Fixtures = append(round.Fixtures, fixture{
				token: cl.Attr(c.parser.fixtures["attr_token"]),
				leftTeam: strings.Split(cl.Attr(c.parser.fixtures["attr_teams"]),
					c.parser.fixtures["attr_teams_delimiter"])[0],
				rightTeam: strings.Split(cl.Attr(c.parser.fixtures["attr_teams"]),
					c.parser.fixtures["attr_teams_delimiter"])[1],
				leftID:  leftID,
				rightID: rightID,
				// Initialise winnerID to -1 to later detect missed predictions,
				// a draw is reflected with a winnerID of 0 and margin of 0
				winnerID: -1,
			})

		})

	})

	// If the login attribute is detected in the response body, authentication has somehow failed
	c.collector.OnHTML(c.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during fixture extraction, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	c.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred during fixture extraction, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during fixture extraction, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	// Client request to the targets fixture endpoint based on the currently active round.
	c.collector.Visit(fmt.Sprint(c.config.urls["fixtures"], round.id))

	return err

}

// setPredictions uses a pre-authenticated client to submit predictions for each fixture to the target.
func (c *client) setPredictions(round *Round) error {

	var err error

	for idx := range round.Fixtures {

		// Should there be no winnderID set for the fixture (i.e. we have missed the prediction somehow),
		// append the fixture details to err using error wrapping (https://golang.org/doc/go1.13#error_wrapping)
		if round.Fixtures[idx].winnerID == -1 {

			if err == nil {
				// No need to wrap err on the first missed prediction
				err = fmt.Errorf("An error has occurred due to a missing predictions, fixture: %s v %s with token %s",
					round.Fixtures[idx].leftTeam,
					round.Fixtures[idx].rightTeam,
					round.Fixtures[idx].token)
			} else {
				err = fmt.Errorf("%w, fixture %s v %s with token %s",
					err, round.Fixtures[idx].leftTeam,
					round.Fixtures[idx].rightTeam,
					round.Fixtures[idx].token)
			}
		} else {
			// Submit parsed prediction query string to target, only token needs escaping at present.
			// This has to be done separately for each fixture (i.e. within the fixture loop) due to the
			// old school AJAX post mechanism used by the target.
			c.collector.Visit(fmt.Sprint(c.config.urls["predictions"],
				fmt.Sprintf(c.parser.predictions["attr_prediction"],
					url.QueryEscape(round.Fixtures[idx].token),
					round.Fixtures[idx].winnerID,
					round.Fixtures[idx].margin,
					round.Fixtures[idx].winnerID,
					round.Fixtures[idx].margin,
					round.Fixtures[idx].winnerID,
					round.Fixtures[idx].margin),
			))
			helpers.Logger.Debugf("Prediction has been submitted, winnerID: %d "+
				"leftTeam: %s, leftID: %d, rightTeam: %s, rightID: %d, margin: %d, token: %s",
				round.Fixtures[idx].winnerID,
				round.Fixtures[idx].leftTeam,
				round.Fixtures[idx].leftID,
				round.Fixtures[idx].rightTeam,
				round.Fixtures[idx].rightID,
				round.Fixtures[idx].margin,
				round.Fixtures[idx].token,
			)
		}
	}

	// If the login attribute is detected in the response body, authentication has somehow failed
	c.collector.OnHTML(c.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during prediction submission, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	c.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred during prediction submission, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during prediction submission, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	return err

}
