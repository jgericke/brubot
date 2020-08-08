package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// getResults uses a pre-authenticated client to retrieve fixture results from a specified round
func (t *Target) getResults(previousRound *PreviousRound) error {

	var err error
	var margin int
	var winner string

	t.Client.collector.OnHTML(t.Client.parser.results["attr_onhtml"], func(e *colly.HTMLElement) {

		e.ForEach(t.Client.parser.results["attr_fixture"], func(_ int, cl *colly.HTMLElement) {

			// Split leftTeam and rightTeam based into array using a known delimeter for
			// team name differentiation. We are assuming that there will always be 2 elements
			// in the array returned from the split, 0 being leftTeam and 1 being rightTeam.
			leftTeam := helpers.CleanName(strings.Split(
				cl.Attr(t.Client.parser.results["attr_t_teams"]),
				t.Client.parser.results["attr_t_teams_delimiter"],
			)[0])
			rightTeam := helpers.CleanName(strings.Split(
				cl.Attr(t.Client.parser.results["attr_t_teams"]),
				t.Client.parser.results["attr_t_teams_delimiter"],
			)[1])
			// If the results parser returns a draw then set margin to 0 and winner to draw
			if strings.EqualFold(cl.ChildText(t.Client.parser.results["attr_t_results"]), t.Client.parser.results["attr_t_draw"]) {
				margin = 0
				winner = "draw"
			} else {
				// Split the winner and margin based on a known delimeter for winner team name
				// and margin
				winner = helpers.CleanName(strings.Split(
					cl.ChildText(t.Client.parser.results["attr_t_results"]),
					t.Client.parser.results["attr_t_winner_delimiter"],
				)[0])
				// Sets and converts margin from string to int
				if marginResult, marginErr := strconv.Atoi(
					strings.Split(
						cl.ChildText(t.Client.parser.results["attr_t_results"]),
						t.Client.parser.results["attr_t_winner_delimiter"])[1],
				); marginErr != nil {
					err = marginErr
				} else {
					margin = marginResult
				}
			}
			// Append extracted result to previousRounds Results
			previousRound.Results = append(previousRound.Results, result{
				leftTeam:  leftTeam,
				rightTeam: rightTeam,
				winner:    winner,
				margin:    margin,
			})

			helpers.Logger.Debugf("Result has been retrieved, leftTeam: %s, rightTeam: %s, winner: %s, margin: %d",
				leftTeam,
				rightTeam,
				winner,
				margin,
			)

		})

	})

	// If the login attribute is detected in the response body, authentication has somehow failed
	t.Client.collector.OnHTML(t.Client.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during results retrieval, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	t.Client.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred results retrieval, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during results retrieval, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	// Client request to the targets results endpoint based on the results roundID
	t.Client.collector.Visit(fmt.Sprint(t.Client.config.urls["results"], previousRound.id))

	return err

}
