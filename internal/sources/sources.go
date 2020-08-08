package sources

import (
	"brubot/config"
)

// Sources holds all predictions extracted for each source
type Sources struct {
	Sources []Source
}

// Source represents a source data location for margin retrieval.
type Source struct {
	Name       string  // Used by reflection call to retrieve margins
	Tournament string  // Tournament name source is providing margins for
	Weight     float64 // Used to calculate aggregated margins based on weighted averages
	Client     client  // Colly client
	Round      Round   // Current round ID
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
			Name:       sourcesConfig.Sources[idx].Name,
			Tournament: sourcesConfig.Sources[idx].Tournament,
			Weight:     sourcesConfig.Sources[idx].Weight,
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
