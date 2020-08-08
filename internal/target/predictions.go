package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// getFixtures uses a pre-authenticated client to extract fixture parameters for a specified round.
func (t *Target) getFixtures(round *Round) error {

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
	t.Client.collector.OnHTML(t.Client.parser.fixtures["attr_onhtml"], func(e *colly.HTMLElement) {

		e.ForEach(t.Client.parser.fixtures["attr_fixture"], func(_ int, cl *colly.HTMLElement) {

			// Convert teamIDs to int (for better living).
			leftID, convErr := strconv.Atoi(cl.Attr(t.Client.parser.fixtures["attr_t_leftid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			rightID, convErr := strconv.Atoi(cl.Attr(t.Client.parser.fixtures["attr_t_rightid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			// Appends a fixture element to a slice of Fixtures
			// within the active Round, setting scraped and parsed fixture
			// parameters.
			round.Fixtures = append(round.Fixtures, fixture{
				token: cl.Attr(t.Client.parser.fixtures["attr_token"]),
				leftTeam: strings.Split(cl.Attr(t.Client.parser.fixtures["attr_teams"]),
					t.Client.parser.fixtures["attr_teams_delimiter"])[0],
				rightTeam: strings.Split(cl.Attr(t.Client.parser.fixtures["attr_teams"]),
					t.Client.parser.fixtures["attr_teams_delimiter"])[1],
				leftID:  leftID,
				rightID: rightID,
				// Initialise winnerID to -1 to later detect missed predictions,
				// a draw is reflected with a winnerID of 0 and margin of 0
				winnerID: -1,
			})

		})

	})

	// If the login attribute is detected in the response body, authentication has somehow failed
	t.Client.collector.OnHTML(t.Client.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during fixture extraction, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	t.Client.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred during fixture extraction, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during fixture extraction, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	// Client request to the targets fixture endpoint based on the currently active round.
	t.Client.collector.Visit(fmt.Sprint(t.Client.config.urls["fixtures"], round.id))

	return err

}

// setPredictions uses a pre-authenticated client to submit predictions for each fixture to the target.
func (t *Target) setPredictions(round *Round) error {

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
			t.Client.collector.Visit(fmt.Sprint(t.Client.config.urls["predictions"],
				fmt.Sprintf(t.Client.parser.predictions["attr_prediction"],
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
	t.Client.collector.OnHTML(t.Client.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during prediction submission, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	t.Client.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred during prediction submission, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during prediction submission, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	return err

}
