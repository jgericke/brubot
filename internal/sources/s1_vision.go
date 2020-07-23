package sources

import (
	"brubot/internal/helpers"
	"fmt"
	"math"
	"strconv"

	"github.com/gocolly/colly/v2"
)

// Vision retrieves predicted margins for the source "Vision"
func (s *Source) Vision() error {

	var err error
	var predictedWinner string
	var tournamentRound int
	// Vision provides margins for multiple tournaments
	tournaments := [2]string{
		s.Client.parser.predictions["attr_t_tournament_1"],
		s.Client.parser.predictions["attr_t_tournament_2"],
	}

	// Client error has occurred attempting .Visit
	s.Client.collector.OnError(func(r *colly.Response, resError error) {
		err = resError
		return
	})
	// debug
	s.Client.collector.OnRequest(func(r *colly.Request) {
		helpers.Logger.Debugf("Prediction retrieval from %s", r.URL.String())
	})

	for _, tournamentName := range tournaments {

		// Second tournament round is not in lock-step with first
		if tournamentName == s.Client.parser.predictions["attr_t_tournament_2"] {
			tournamentRound = s.Round.id - 3
		} else {
			tournamentRound = s.Round.id
		}

		// Collect margin predictions from source
		// This will always be bespoke for each source
		s.Client.collector.OnHTML(fmt.Sprintf(s.Client.parser.predictions["attr_onhtml"], tournamentName, tournamentRound), func(e *colly.HTMLElement) {

			e.ForEach("tr", func(_ int, el *colly.HTMLElement) {

				leftTeam := el.ChildText(s.Client.parser.predictions["attr_t_leftteam"])
				rightTeam := el.ChildText(s.Client.parser.predictions["attr_t_rightteam"])

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

	}

	s.Client.collector.Visit(s.Client.config.urls["predictions"])

	return err

}
