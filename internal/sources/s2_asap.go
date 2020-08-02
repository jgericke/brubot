package sources

import (
	"brubot/internal/helpers"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

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
			leftTeam = strings.Replace(strings.ToLower(leftTeam), "the ", "", -1)
			rightTeam = strings.Replace(strings.ToLower(rightTeam), "the ", "", -1)

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
