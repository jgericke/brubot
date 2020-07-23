package target

import (
	"brubot/config"
	"brubot/internal/helpers"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Endpoint represents a target for prediction submission
type Endpoint struct {
	Round  Round  // Round ID, fixtures and predictions for a specific found
	Auth   auth   // Client authentication cookie
	Client client // Colly client instance
}

// Round contains all fixtures and associated prediction per fixture
type Round struct {
	id          int        // ID for a current round, determined by date
	Fixtures    []fixture  // The fixutes (matches) within a round/round ID
	Predictions prediction // Predictions associated to each fixture
}

// Winning team name and margin/point deficit for a fixture prediction
type prediction struct {
	teamMargin map[string]int // winner: margin
}

// Represents all parameters per-fixture
type fixture struct {
	token     string // Unique fixture token, extracted from target
	leftTeam  string // teamA
	rightTeam string // teamB
	leftID    int    // Unique identifer for teamA, extracted from target
	rightID   int    // Unique identifer for teamB, extracted from target
	winnerID  int    // Set to teamA or teamB identifer based on prediction
	margin    int    // Point difference for winning team based on prediction
}

// Init sets an Endpoint up with global and target specific configuration paramaeters.
func (e *Endpoint) Init(globalConfig config.GlobalConfig, targetConfig config.TargetConfig) {

	// Target authentication establishes successful auth, populates a cookiejar with auth
	// token(s) to set on client for subsequent querying.
	//
	// Parameters from config.TargetConfig.Auth passed to internal/target/auth.go -> auth
	e.Auth = auth{
		url:            targetConfig.Auth.URL,
		parameters:     targetConfig.Auth.Parameters,
		passwordEncode: targetConfig.Auth.PasswordEncode,
		method:         targetConfig.Auth.Method,
		errorMsg:       targetConfig.Auth.ErrorMsg,
		timeout:        targetConfig.Auth.Timeout,
		headers:        targetConfig.Auth.Headers,
	}
	// Colly client settings for querying target fixtures and setting predictions.
	//
	// Parameters from config.TargetConfig.Client passed to internal/target/client.go -> clientConfig
	e.Client = client{
		config: clientConfig{
			urls:                targetConfig.Client.URLs,
			ignoreRobots:        targetConfig.Client.IgnoreRobots,
			enableCache:         targetConfig.Client.EnableCache,
			cacheDir:            targetConfig.Client.CacheDir,
			dialTimeout:         targetConfig.Client.DialTimeout,
			tlsHandShakeTimeout: targetConfig.Client.TLSHandShakeTimeout,
		},
		// Parameters from config.TargetConfig.Client.Parser passed tointernal/target/client.go -> clientParser
		parser: clientParser{
			login:       targetConfig.Client.Parser.Login,
			fixtures:    targetConfig.Client.Parser.Fixtures,
			predictions: targetConfig.Client.Parser.Predictions,
		},
	}

	// Globals allow easier parameter setting across multiple http clients
	//
	// At present only user agent can be set globally.
	if targetConfig.UseGlobals {
		e.Auth.userAgent = globalConfig.UserAgent
		e.Client.config.userAgent = globalConfig.UserAgent
	} else {
		e.Auth.userAgent = targetConfig.Auth.UserAgent
		e.Client.config.userAgent = targetConfig.Client.UserAgent
	}

}

// Authenticate builds and sends auth string to target and populates
// a cookiejar to be passed to colly on successful auth.
func (e *Endpoint) Authenticate() error {

	// Call to authenticate method, results in population of auth token
	// within cookiejar
	if err := e.Auth.authenticate(e.Auth.timeout); err != nil {
		return err
	}
	// Initialises client with all client specific parameters, passing
	// auth cookie jar for authenticating subsequent queries.
	if err := e.Client.init(e.Auth.cookieJar); err != nil {
		return err
	}

	return nil

}

// Fixtures retrieves all fixtures details within a round based on roundID
// and populates Round.Fixtures
func (e *Endpoint) Fixtures(roundID int) error {

	// roundID is determined by the current date within
	// preset fixtures date range at time of execution
	e.Round.id = roundID

	if err := e.Client.getFixtures(&e.Round); err != nil {
		return err
	}

	return nil

}

// Predictions handles mapping predictions to fixtures, sets winnerID and margin fields
// for matched fixtures and calls client with predictions for submission to target.
func (e *Endpoint) Predictions(predictions map[string]int) error {

	// predictions are expected to be in the format winningTeamName: margin
	for team, margin := range predictions {

		// Strip out any article and trailing whitespace from team name.
		team = strings.Replace(strings.ToLower(team), "the ", "", -1)

		for idx := range e.Round.Fixtures {
			// fuzzy.RankMatchNormalizedFold provides string matching with Unicode normalisation,
			// where 0 is an exact match, and greater than 0 means less matching characters at higher values.
			// Naively using a scoring of 0 or greater as team name matching criteria.

			// Sets Fixture teamID as the winnerID and margin when either a left or right
			// team in the fixture matches with the predictions's winning team.
			// TeamIDs are retrieved from the target and are randomish/too inconsistent to map up front.

			if fuzzy.RankMatchNormalizedFold(team, e.Round.Fixtures[idx].leftTeam) >= 0 {
				if margin == 0 {
					// Indicates fixture prediction is a draw (margin = 0 / winner_id = 0)
					e.Round.Fixtures[idx].winnerID = 0
				} else {
					e.Round.Fixtures[idx].winnerID = e.Round.Fixtures[idx].leftID
				}
				e.Round.Fixtures[idx].margin = margin
				helpers.Logger.Debugf("Prediction has been set: leftTeam: %s winnerID: %d margin: %d, token: %s",
					e.Round.Fixtures[idx].leftTeam,
					e.Round.Fixtures[idx].winnerID,
					e.Round.Fixtures[idx].margin,
					e.Round.Fixtures[idx].token,
				)
				break
			}

			if fuzzy.RankMatchNormalizedFold(team, e.Round.Fixtures[idx].rightTeam) >= 0 {
				if margin == 0 {
					e.Round.Fixtures[idx].winnerID = 0
				} else {
					e.Round.Fixtures[idx].winnerID = e.Round.Fixtures[idx].rightID
				}

				e.Round.Fixtures[idx].margin = margin
				helpers.Logger.Debugf("Prediction has been set: rightTeam: %s winnerID: %d margin: %d, token: %s",
					e.Round.Fixtures[idx].rightTeam,
					e.Round.Fixtures[idx].winnerID,
					e.Round.Fixtures[idx].margin,
					e.Round.Fixtures[idx].token,
				)
				break
			}

		}
	}

	// Call to client to set matched predictions for each fixture
	if err := e.Client.setPredictions(&e.Round); err != nil {
		return err
	}

	return nil

}
