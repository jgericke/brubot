package target

import (
	"brubot/config"
)

// Endpoint target for predictions
type Endpoint struct {
	Round      Round
	Auth       auth
	Client     client
	TargetURLs targetURLs
}

// Round contains endpoint fixtures
type Round struct {
	id       int
	Fixtures []fixture
}

type fixture struct {
	token     string
	leftTeam  string
	rightTeam string
	leftID    int
	rightID   int
	winnerID  int
	margin    int
}

type targetURLs struct {
	extractURL string
}

// Init target endpoint
func (e *Endpoint) Init(globalConfig config.GlobalConfig, targetConfig config.TargetConfig, roundID int) {

	e.Round.id = roundID
	e.Auth = auth{
		url:            targetConfig.Auth.URL,
		parameters:     targetConfig.Auth.Parameters,
		passwordEncode: targetConfig.Auth.PasswordEncode,
		method:         targetConfig.Auth.Method,
		errorMsg:       targetConfig.Auth.ErrorMsg,
		timeout:        targetConfig.Auth.Timeout,
		headers:        targetConfig.Auth.Headers,
	}
	e.Client = client{
		config: clientConfig{
			urls:                targetConfig.Client.URLs,
			ignoreRobots:        targetConfig.Client.IgnoreRobots,
			enableCache:         targetConfig.Client.EnableCache,
			cacheDir:            targetConfig.Client.CacheDir,
			dialTimeout:         targetConfig.Client.DialTimeout,
			tlsHandShakeTimeout: targetConfig.Client.TLSHandShakeTimeout,
		},
		parser: clientParser{
			login:    targetConfig.Client.Parser.Login,
			fixtures: targetConfig.Client.Parser.Fixtures,
		},
	}

	if targetConfig.UseGlobals {
		e.Auth.userAgent = globalConfig.UserAgent
		e.Client.config.userAgent = globalConfig.UserAgent
	} else {
		e.Auth.userAgent = targetConfig.Auth.UserAgent
		e.Client.config.userAgent = targetConfig.Client.UserAgent
	}

}

// Fixtures extracts fixtures for a given roundID
func (e *Endpoint) Fixtures() error {

	/*
		if err := e.Auth.authenticate(e.Auth.timeout); err != nil {
			return err
		}
	*/

	e.Client.init(e.Auth.cookieJar)
	if err := e.Client.getFixtures(&e.Round); err != nil {
		return err
	}

	return nil

}
