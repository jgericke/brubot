package target

import (
	"brubot/internal/helpers"
	"errors"
	"fmt"
	"net/url"

	"github.com/gocolly/colly/v2"
)

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
