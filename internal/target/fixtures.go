package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Fixtures retrieves all fixtures details within a round based on roundID
// and populates Round.Fixtures
func (t *Target) Fixtures(roundID int) error {

	// roundID is determined by the current date within
	// preset fixtures date range at time of execution
	t.Round.id = roundID

	if err := t.getFixtures(); err != nil {
		return err
	}

	return nil

}

// getFixtures uses a pre-authenticated client to extract fixture parameters for a specified round.
func (t *Target) getFixtures() error {

	var err error

	// Scrapes and parses fixtures for the active round (set via t.Round.id).
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

			// Set token from parsed token attribute
			token := cl.Attr(t.Client.parser.fixtures["attr_token"])
			// Split left and right teams using known delimeter, assumes
			// the array returned will consist of only 2 elements being respsective team names
			leftTeam := helpers.CleanName(strings.Split(cl.Attr(t.Client.parser.fixtures["attr_teams"]),
				t.Client.parser.fixtures["attr_teams_delimiter"])[0])
			rightTeam := helpers.CleanName(strings.Split(cl.Attr(t.Client.parser.fixtures["attr_teams"]),
				t.Client.parser.fixtures["attr_teams_delimiter"])[1])

			// Appends a fixture element to a slice of Fixtures
			// within the active Round, setting scraped and parsed fixture
			// parameters.
			t.Round.Fixtures = append(t.Round.Fixtures, fixture{
				token:     token,
				leftTeam:  leftTeam,
				rightTeam: rightTeam,
				leftID:    leftID,
				rightID:   rightID,
				// Initialise winnerID to -1 to later detect missed predictions,
				// a draw is reflected with a winnerID of 0 and margin of 0
				winnerID: -1,
			})

			helpers.Logger.Debugf("Fixture has been retrieved for round: %d, token: %s leftTeam: %s (id: %d), rightTeam: %s (id %d)",
				t.Round.id,
				token,
				leftTeam,
				leftID,
				rightTeam,
				rightID,
			)

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
	t.Client.collector.Visit(fmt.Sprint(t.Client.config.urls["fixtures"], t.Round.id))

	return err

}
