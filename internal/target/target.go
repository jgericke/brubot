package target

import (
	"brubot/config"
)

// Target is everything required to submit a prediction
type Target struct {
	Round         Round         // Round ID, fixtures and predictions for a specific found
	PreviousRound PreviousRound // Round ID and results for the previous round of fixtures
	Auth          auth          // Client authentication cookie
	Client        client        // Colly client instance
}

// Round contains all fixtures and associated prediction per fixture
type Round struct {
	id          int        // ID for a current round, determined by date
	Fixtures    []fixture  // The fixutes (matches) within a round/round ID
	Predictions prediction // Predictions associated to each fixture
}

// PreviousRound contains fixture results for the most recent completed round
type PreviousRound struct {
	id      int      // Will generally be Round.id - 1
	Results []result // Previous rounds fixtures with match outcomes
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

// Result of a completed fixture (similar to fixture but *Different*)
type result struct {
	leftTeam  string // teamA
	rightTeam string // teamB
	winner    string // Set to teamA or teamB identifer based on fixture results (or 'draw' in a draw)
	margin    int    // Point difference for winning team based / winning margin
}

// Init sets a Target up with global and target specific configuration paramaeters.
func (t *Target) Init(globalConfig config.GlobalConfig, targetConfig config.TargetConfig) {

	// Target authentication establishes successful auth, populates a cookiejar with auth
	// token(s) to set on client for subsequent querying.
	//
	// Parameters from config.TargetConfig.Auth passed to internal/target/auth.go -> auth
	t.Auth = auth{
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
	t.Client = client{
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
			results:     targetConfig.Client.Parser.Results,
			predictions: targetConfig.Client.Parser.Predictions,
		},
	}

	// Globals allow easier parameter setting across multiple http clients
	//
	// At present only user agent can be set globally.
	if targetConfig.UseGlobals {
		t.Auth.userAgent = globalConfig.UserAgent
		t.Client.config.userAgent = globalConfig.UserAgent
	} else {
		t.Auth.userAgent = targetConfig.Auth.UserAgent
		t.Client.config.userAgent = targetConfig.Client.UserAgent
	}

}
