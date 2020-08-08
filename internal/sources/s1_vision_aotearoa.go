package sources

import (
	"brubot/internal/helpers"
	"fmt"
	"math"
	"strconv"

	"github.com/gocolly/colly/v2"
)

// VisionAotearoa retrieves predicted margins for the source "VisionAotearoa"
func (s *Source) VisionAotearoa() error {

	var err error
	var predictedWinner string

	// Client error has occurred attempting .Visit
	s.Client.collector.OnError(func(r *colly.Response, resError error) {
		err = resError
		return
	})
	// debug
	s.Client.collector.OnRequest(func(r *colly.Request) {
		helpers.Logger.Debugf("Prediction retrieval from %s", r.URL.String())
	})

	// Collect margin predictions from source
	// This will always be bespoke for each source
	s.Client.collector.OnHTML(fmt.Sprintf(s.Client.parser.predictions["attr_onhtml"], s.Round.id), func(e *colly.HTMLElement) {

		e.ForEach(s.Client.parser.predictions["attr_t_iterator"], func(_ int, el *colly.HTMLElement) {

			leftTeam := el.ChildText(s.Client.parser.predictions["attr_t_leftteam"])
			rightTeam := el.ChildText(s.Client.parser.predictions["attr_t_rightteam"])
			// Clean up team names for easier matching
			leftTeam = helpers.CleanName(leftTeam)
			rightTeam = helpers.CleanName(rightTeam)

			if predictedMargin, marginErr := strconv.Atoi(el.ChildText(s.Client.parser.predictions["attr_t_margin"])); marginErr != nil {
				err = marginErr
			} else {
				if predictedMargin > 0 {
					predictedWinner = leftTeam
				} else {
					// Vision reflects a leftTeam loss with a negative int,
					// we flip that with to positive and set the winner to rightTeam
					predictedMargin = int(math.Abs(float64(predictedMargin)))
					predictedWinner = rightTeam
				}
				s.Round.Fixtures = append(s.Round.Fixtures, fixture{
					leftTeam:  leftTeam,
					rightTeam: rightTeam,
					winner:    predictedWinner,
					margin:    predictedMargin,
				})

			}
		})
	})

	s.Client.collector.Visit(s.Client.config.urls["predictions"])

	return err

}
