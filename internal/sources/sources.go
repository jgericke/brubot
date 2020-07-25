package sources

import (
	"brubot/config"
	"brubot/internal/helpers"
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/lib/pq"
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

		// Call prediction extraction method using the Source objects Name property,
		// i.e. if the source is named foo, a method should exist of the same name for
		// retrieving predictions
		errv := reflect.ValueOf(&s.Sources[idx]).MethodByName(s.Sources[idx].Name).Call([]reflect.Value{})

		// error checking is different here as valueof returns a reflected interface
		if !errv[0].IsNil() {
			// Use error wrapping to collect errors per-source but not fail outright for all
			// prediction retrievals (faily messy)
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

// Update records extracted source predictions for a round
func (s *Sources) Update(db *sql.DB) error {

	// Temporary ID for duplicate prediction checking
	var tmpID int
	// Create an empty context for update transaction
	sqlCtx := context.Background()
	// Start sql transaction
	sqlTxn, err := db.BeginTx(sqlCtx, nil)
	if err != nil {
		return err
	}
	// prepare sql statement with COPY FROM
	sqlStmt, err := sqlTxn.Prepare(pq.CopyIn("predictions", "round_id", "source", "leftteam", "rightteam", "winner", "margin"))
	if err != nil {
		return err
	}

	helpers.Logger.Debug("Prediction update is emminent, hold tight...")

	for idx := range s.Sources {
		for f := range s.Sources[idx].Round.Fixtures {

			// Ugly check to establish if a source already has a prediction recorded
			// against the round id and fixture parameters
			// Kinda breaking the rules when calling a db method with a
			// transaction underway
			sqlRow := db.QueryRowContext(sqlCtx,
				"SELECT id FROM predictions WHERE round_id=$1"+
					"AND source=$2 AND leftteam=$3 AND rightteam=$4"+
					"AND winner=$5 AND margin=$6",
				s.Sources[idx].Round.id,
				s.Sources[idx].Name,
				s.Sources[idx].Round.Fixtures[f].leftTeam,
				s.Sources[idx].Round.Fixtures[f].rightTeam,
				s.Sources[idx].Round.Fixtures[f].winner,
				s.Sources[idx].Round.Fixtures[f].margin).Scan(&tmpID)
			switch {
			case sqlRow == sql.ErrNoRows:
				// ErrNoRows means we are good to go, execute CopyIn
				// with round id and fixture parameters
				_, err = sqlStmt.Exec(
					s.Sources[idx].Round.id,
					s.Sources[idx].Name,
					s.Sources[idx].Round.Fixtures[f].leftTeam,
					s.Sources[idx].Round.Fixtures[f].rightTeam,
					s.Sources[idx].Round.Fixtures[f].winner,
					s.Sources[idx].Round.Fixtures[f].margin,
				)
				if err != nil {
					return err
				}
			case sqlRow != nil:
				// Error occurred during query
				return sqlRow
			default:
				helpers.Logger.Debugf("Prediction update omitted as record already exists with ID: %d", tmpID)
			}

		}
	}

	err = sqlStmt.Close()
	if err != nil {
		return err
	}

	err = sqlTxn.Commit()
	if err != nil {
		return err
	}

	helpers.Logger.Debug("Prediction update completed sans incidents")

	return nil
}
