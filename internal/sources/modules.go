/*
	The source "modules", providing margin predictions for upcoming fixutres.
*/
package sources

import (
	"brubot/internal/helpers"
	"fmt"
	"math"
	"regexp"
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

// VisionAu retrieves predicted margins for the source "VisionAu"
func (s *Source) VisionAu() error {

	var err error
	var predictedWinner string
	var auRoundID = s.Round.id - 3 // Tournament is not in lock-step round-wise

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
	s.Client.collector.OnHTML(fmt.Sprintf(s.Client.parser.predictions["attr_onhtml"], auRoundID), func(e *colly.HTMLElement) {

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

// Asap retrieves predicted margins for the source "Asap"
func (s *Source) Asap() error {

	var err, marginErr error
	var predictedWinner string
	var predictedMargin int
	re := regexp.MustCompile("[0-9]+")

	// client error has occurred attempting .Visit
	s.Client.collector.OnError(func(r *colly.Response, resError error) {
		err = resError
		return
	})
	// debug
	s.Client.collector.OnRequest(func(r *colly.Request) {
		helpers.Logger.Debugf("Prediction retrieval from %s", r.URL.String())
	})

	s.Client.collector.OnHTML(s.Client.parser.predictions["attr_onhtml"], func(e *colly.HTMLElement) {

		e.ForEach(s.Client.parser.predictions["attr_t_iterator"], func(_ int, el *colly.HTMLElement) {

			leftTeam := e.ChildText(s.Client.parser.predictions["attr_t_leftteam"])
			leftTeamMargin := re.FindAllString(e.ChildText(s.Client.parser.predictions["attr_t_leftmargin"]), -1)

			rightTeam := e.ChildText(s.Client.parser.predictions["attr_t_rightteam"])
			rightTeamMargin := re.FindAllString(e.ChildText(s.Client.parser.predictions["attr_t_rightmargin"]), -1)

			// Clean up team names for easier matching
			leftTeam = helpers.CleanName(leftTeam)
			rightTeam = helpers.CleanName(rightTeam)

			// There should only be a single element in these margin results (hopefully!)
			if len(leftTeamMargin) > 0 {
				predictedWinner = leftTeam
				if predictedMargin, marginErr = strconv.Atoi(leftTeamMargin[0]); marginErr != nil {
					err = marginErr
				}

			}
			if len(rightTeamMargin) > 0 {
				predictedWinner = rightTeam
				if predictedMargin, marginErr = strconv.Atoi(rightTeamMargin[0]); marginErr != nil {
					err = marginErr
				}
			}
			s.Round.Fixtures = append(s.Round.Fixtures, fixture{
				leftTeam:  leftTeam,
				rightTeam: rightTeam,
				winner:    predictedWinner,
				margin:    predictedMargin,
			})

		})

	})

	s.Client.collector.Visit(s.Client.config.urls["predictions"])

	return err

}
