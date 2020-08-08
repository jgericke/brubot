package sources

import (
	"brubot/internal/helpers"
	"fmt"
)

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
