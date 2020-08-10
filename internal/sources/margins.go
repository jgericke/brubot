package sources

import (
	"brubot/internal/helpers"
	"database/sql"
	"fmt"
	"math"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Weights calculates the moving weight a source should be allocated using the sources aggregate prediction accuracy
func (s *Sources) Weights(previousRoundID int, db *sql.DB) error {

	var source, leftTeam, rightTeam, winner string
	var margin, marginDifference int
	var marginError float64

	previousRound := new(Round)
	previousRound.id = previousRoundID

	// First up, retrieve last rounds results
	txn, err := db.Begin()
	if err != nil {
		return err
	}

	sqlResults, err := txn.Query("select leftteam, rightteam, winner, margin from results where round_id=$1", previousRoundID)
	if err != nil {
		return err
	}

	for sqlResults.Next() {

		err := sqlResults.Scan(&leftTeam, &rightTeam, &winner, &margin)
		if err != nil {
			return err
		}

		// Append results as fixtures of previousRound
		previousRound.Fixtures = append(previousRound.Fixtures, fixture{
			leftTeam:  leftTeam,
			rightTeam: rightTeam,
			winner:    winner,
			margin:    margin,
		})

	}

	// Next up retrieve last rounds predictions per-source and append to a temporary Source instance
	sqlPredictions, err := txn.Query("select source, leftteam, rightteam, winner, margin from predictions where round_id=$1", previousRoundID)
	if err != nil {
		return err
	}

	var tmpSource []Source
	var tmpFixture []fixture

	for sqlPredictions.Next() {

		err := sqlPredictions.Scan(&source, &leftTeam, &rightTeam, &winner, &margin)
		if err != nil {
			return err
		}

		tmpSource = append(tmpSource, Source{
			Name: source,
			Round: Round{
				id: previousRoundID,
				Fixtures: append(tmpFixture, fixture{
					leftTeam:  leftTeam,
					rightTeam: rightTeam,
					winner:    winner,
					margin:    margin,
				}),
			},
		})
	}

	for sIdx := range tmpSource {
		for rIdx := range previousRound.Fixtures {
			for f := range tmpSource[sIdx].Round.Fixtures {

				if fuzzy.RankMatchNormalizedFold(tmpSource[sIdx].Round.Fixtures[f].leftTeam, previousRound.Fixtures[rIdx].leftTeam) >= 0 &&
					fuzzy.RankMatchNormalizedFold(tmpSource[sIdx].Round.Fixtures[f].rightTeam, previousRound.Fixtures[rIdx].rightTeam) >= 0 {

					// Confirm prediction had the correct winning team:
					if fuzzy.RankMatchNormalizedFold(tmpSource[sIdx].Round.Fixtures[f].winner, previousRound.Fixtures[rIdx].winner) >= 0 {
						// Calc difference between predicted margin and actual result margin as marginDifference
						switch {
						case tmpSource[sIdx].Round.Fixtures[f].margin == previousRound.Fixtures[rIdx].margin:
							marginDifference = 0
							marginError = 0
						default:
							marginDifference = int(math.Abs(float64(previousRound.Fixtures[rIdx].margin - tmpSource[sIdx].Round.Fixtures[f].margin)))
							marginError = (float64(marginDifference) / float64(previousRound.Fixtures[rIdx].margin)) * 100
						}

						helpers.Logger.Infof("tmpSource matched! source name: %s\npred winner: %s result winner: %s\n"+
							"pred margin: %d result margin: %d (Prediction off by: %d percent error: %f)",
							tmpSource[sIdx].Name,
							tmpSource[sIdx].Round.Fixtures[f].winner,
							previousRound.Fixtures[rIdx].winner,
							tmpSource[sIdx].Round.Fixtures[f].margin,
							previousRound.Fixtures[rIdx].margin,
							marginDifference,
							marginError,
						)

					} else {

						helpers.Logger.Infof("tmpSource matched but predicted the wrong winner! source name: %s\npred winner: %s result winner: %s\n"+
							"pred margin: %d result margin: %d",
							tmpSource[sIdx].Name,
							tmpSource[sIdx].Round.Fixtures[f].winner,
							previousRound.Fixtures[rIdx].winner,
							tmpSource[sIdx].Round.Fixtures[f].margin,
							previousRound.Fixtures[rIdx].margin,
						)

					}

				}

			}
		}
	}

	return nil

}

// Margins figures out the best margins in town from retrievd predictions and previous results
func (s *Sources) Margins(roundID int) (map[string]int, error) {

	var err error
	var margins map[string]int
	margins = make(map[string]int)

	// In each source find fixtures with matching *winners* (for now, assuming only one team can win one fixture per round),
	// and calculate weighted margins where applicable
	for idx := range s.Sources {
		for f := range s.Sources[idx].Round.Fixtures {
			if _, ok := margins[s.Sources[idx].Round.Fixtures[f].winner]; ok {
				if s.Sources[idx].Weight != 0 {
					// Add new weighted margin to predictions existing aggregate weighted margin
					margins[s.Sources[idx].Round.Fixtures[f].winner] = margins[s.Sources[idx].Round.Fixtures[f].winner] + int(float64(s.Sources[idx].Round.Fixtures[f].margin)*s.Sources[idx].Weight)
					helpers.Logger.Debugf("Margin updated from source: %s, winner: %s, weigthed margin now: %d",
						s.Sources[idx].Name,
						s.Sources[idx].Round.Fixtures[f].winner,
						margins[s.Sources[idx].Round.Fixtures[f].winner],
					)
				} else {
					// Matched margin prediction without a weighted source, implication being there are 2 sources with the same
					// tournament that should have margins aggregated using weighted averages but dont have weights
					helpers.Logger.Errorf("Margin matched from new source without a weight, source: %s", s.Sources[idx].Name)
					if err == nil {
						err = fmt.Errorf("Margin matched from new source without a weight, source: %s", s.Sources[idx].Name)
					} else {
						err = fmt.Errorf("%w, Margin matched from new source without a weight, source: %s", err, s.Sources[idx].Name)
					}
				}
			} else {
				// Add new margin for sources with weight specified (calculating weighted average)
				if s.Sources[idx].Weight != 0 {
					margins[s.Sources[idx].Round.Fixtures[f].winner] = int(float64(s.Sources[idx].Round.Fixtures[f].margin) * s.Sources[idx].Weight)
					helpers.Logger.Debugf("Margin with weighted prediction added from source: %s, winner: %s, weighted margin: %d",
						s.Sources[idx].Name,
						s.Sources[idx].Round.Fixtures[f].winner,
						margins[s.Sources[idx].Round.Fixtures[f].winner],
					)
				} else {
					// Add new margin for sources without weight specified (source margin is our margin)
					margins[s.Sources[idx].Round.Fixtures[f].winner] = s.Sources[idx].Round.Fixtures[f].margin
					helpers.Logger.Debugf("Margin without weighted prediction added from source: %s, winner: %s, margin: %d",
						s.Sources[idx].Name,
						s.Sources[idx].Round.Fixtures[f].winner,
						margins[s.Sources[idx].Round.Fixtures[f].winner],
					)
				}
			}
		}

	}

	return margins, err

}
