package sources

import (
	"brubot/config"
	"brubot/internal/helpers"
	"fmt"
	"reflect"
)

// Sources holds all predictions extracted for each source
type Sources struct {
	Sources []Source
}

// Source represents a source data location for margin retrieval.
type Source struct {
	Name   string
	Client client
	Round  Round
}

// Round contains all fixtures and associated prediction per fixture
// and attempts to mirror target Round for easy translation
type Round struct {
	id       int       // ID for a current round, determined by date
	Fixtures []fixture // All fixutes (matches) within a round/round ID
}

// fixture represents a match within a round
// and attempts to mirror target fixtures for easy translation
type fixture struct {
	leftTeam  string // teamA
	rightTeam string // teamB
	winner    string // team name of predicted winning team
	margin    int    // Point difference for winning team based on prediction
}

// Init builds Sources by iterating through all configured source endpoints within
// config.SourcesConfig and creating a slice element for each with relevant
// configurables set.
func (s *Sources) Init(globalConfig config.GlobalConfig, sourcesConfig config.SourcesConfig) {

	for idx := range sourcesConfig.Sources {

		s.Sources = append(s.Sources, Source{
			Name: sourcesConfig.Sources[idx].Name,
			Client: client{
				config: clientConfig{
					urls:                sourcesConfig.Sources[idx].Client.URLs,
					ignoreRobots:        sourcesConfig.Sources[idx].Client.IgnoreRobots,
					enableCache:         sourcesConfig.Sources[idx].Client.EnableCache,
					cacheDir:            sourcesConfig.Sources[idx].Client.CacheDir,
					dialTimeout:         sourcesConfig.Sources[idx].Client.DialTimeout,
					tlsHandShakeTimeout: sourcesConfig.Sources[idx].Client.TLSHandShakeTimeout,
				},
				parser: clientParser{
					predictions: sourcesConfig.Sources[idx].Client.Parser.Predictions,
				},
			},
		})

		// Set global parameters where applicable
		if sourcesConfig.Sources[idx].UseGlobals {
			s.Sources[idx].Client.config.userAgent = globalConfig.UserAgent
		} else {
			s.Sources[idx].Client.config.userAgent = sourcesConfig.Sources[idx].Client.UserAgent
		}
		// Go ahead an initialise each source endpoints colly client
		// while initialising the source itself (saves time and money)
		s.Sources[idx].Client.init()
	}

}

// Predictions iterates through all sources and uses reflection to call
// each sources corresponding method by name, which in turn populates each source
// with predictions per fixture
func (s *Sources) Predictions(roundID int) error {

	var err error

	for idx := range s.Sources {

		s.Sources[idx].Round.id = roundID

		// Call the Source objects method by method name using the Source objects Name property
		errv := reflect.ValueOf(&s.Sources[idx]).MethodByName(s.Sources[idx].Name).Call([]reflect.Value{})

		// error checking is different as valueof returns a reflected interface
		if !errv[0].IsNil() {
			// Use error wrapping to collect errors per-source but not fail outright for all
			// prediction retrievals
			if err == nil {
				err = fmt.Errorf("Failed source: %s error: %v", s.Sources[idx].Name, errv[0].Interface().(error))
			} else {
				err = fmt.Errorf("%w, Failed source: %s error: %v", err, s.Sources[idx].Name, errv[0].Interface().(error))
			}
		}

		for f := range s.Sources[idx].Round.Fixtures {
			helpers.Logger.Debugf("Prediction has been retrieved from: %s letfTeam: %s rightTeam: %s, winner: %s, margin %d",
				s.Sources[idx].Name,
				s.Sources[idx].Round.Fixtures[f].leftTeam,
				s.Sources[idx].Round.Fixtures[f].rightTeam,
				s.Sources[idx].Round.Fixtures[f].winner,
				s.Sources[idx].Round.Fixtures[f].margin,
			)

		}

	}

	return err

}
