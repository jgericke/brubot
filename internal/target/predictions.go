package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Predictions handles mapping predictions to fixtures, sets winnerID and margin fields
// for matched fixtures and calls client with predictions for submission to target.
func (t *Target) Predictions(predictions map[string]int) error {

	// predictions are expected to be in the format winningTeamName: margin
	for team, margin := range predictions {

		// Strip out any article and trailing whitespace from team name
		team = helpers.CleanName(team)

		for idx := range t.Round.Fixtures {
			// fuzzy.RankMatchNormalizedFold provides string matching with Unicode normalisation,
			// where 0 is an exact match, and greater than 0 means less matching characters at higher values.
			// Naively using a scoring of 0 or greater as team name matching criteria.

			// Sets Fixture teamID as the winnerID and margin when either a left or right
			// team in the fixture matches with the predictions's winning team.
			// TeamIDs are retrieved from the target and are randomish/too inconsistent to map up front.
			if fuzzy.RankMatchNormalizedFold(team, t.Round.Fixtures[idx].leftTeam) >= 0 {
				if margin == 0 {
					// Indicates fixture prediction is a draw (margin = 0 / winner_id = 0)
					t.Round.Fixtures[idx].winnerID = 0
				} else {
					t.Round.Fixtures[idx].winnerID = t.Round.Fixtures[idx].leftID
				}

				t.Round.Fixtures[idx].margin = margin
				helpers.Logger.Debugf("Prediction has been set: leftTeam: %s winnerID: %d margin: %d, token: %s",
					t.Round.Fixtures[idx].leftTeam,
					t.Round.Fixtures[idx].winnerID,
					t.Round.Fixtures[idx].margin,
					t.Round.Fixtures[idx].token,
				)
				break
			}

			if fuzzy.RankMatchNormalizedFold(team, t.Round.Fixtures[idx].rightTeam) >= 0 {
				if margin == 0 {
					t.Round.Fixtures[idx].winnerID = 0
				} else {
					t.Round.Fixtures[idx].winnerID = t.Round.Fixtures[idx].rightID
				}

				t.Round.Fixtures[idx].margin = margin
				helpers.Logger.Debugf("Prediction has been set: rightTeam: %s winnerID: %d margin: %d, token: %s",
					t.Round.Fixtures[idx].rightTeam,
					t.Round.Fixtures[idx].winnerID,
					t.Round.Fixtures[idx].margin,
					t.Round.Fixtures[idx].token,
				)
				break
			}

		}
	}

	// Call to client to set matched predictions for each fixture
	if err := t.setPredictions(); err != nil {
		return err
	}

	return nil

}

// setPredictions uses a pre-authenticated client to submit predictions for each fixture to the target.
func (t *Target) setPredictions() error {

	var err error

	for idx := range t.Round.Fixtures {

		// Should there be no winnderID set for the fixture (i.e. we have missed the prediction somehow),
		// append the fixture details to err using error wrapping (https://golang.org/doc/go1.13#error_wrapping)
		if t.Round.Fixtures[idx].winnerID == -1 {

			if err == nil {
				// No need to wrap err on the first missed prediction
				err = fmt.Errorf("An error has occurred due to a missing predictions, fixture: %s v %s with token %s",
					t.Round.Fixtures[idx].leftTeam,
					t.Round.Fixtures[idx].rightTeam,
					t.Round.Fixtures[idx].token)
			} else {
				err = fmt.Errorf("%w, fixture %s v %s with token %s",
					err, t.Round.Fixtures[idx].leftTeam,
					t.Round.Fixtures[idx].rightTeam,
					t.Round.Fixtures[idx].token)
			}
		} else {
			// Submit parsed prediction query string to target, only token needs escaping at present.
			// This has to be done separately for each fixture (i.e. within the fixture loop) due to the
			// old school AJAX post mechanism used by the target.
			t.Client.collector.Visit(fmt.Sprint(t.Client.config.urls["predictions"],
				fmt.Sprintf(t.Client.parser.predictions["attr_prediction"],
					url.QueryEscape(t.Round.Fixtures[idx].token),
					t.Round.Fixtures[idx].winnerID,
					t.Round.Fixtures[idx].margin,
					t.Round.Fixtures[idx].winnerID,
					t.Round.Fixtures[idx].margin,
					t.Round.Fixtures[idx].winnerID,
					t.Round.Fixtures[idx].margin),
			))
			helpers.Logger.Debugf("Prediction has been submitted for round: %d, winnerID: %d, "+
				"leftTeam: %s, leftID: %d, rightTeam: %s, rightID: %d, margin: %d, token: %s",
				t.Round.id,
				t.Round.Fixtures[idx].winnerID,
				t.Round.Fixtures[idx].leftTeam,
				t.Round.Fixtures[idx].leftID,
				t.Round.Fixtures[idx].rightTeam,
				t.Round.Fixtures[idx].rightID,
				t.Round.Fixtures[idx].margin,
				t.Round.Fixtures[idx].token,
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
