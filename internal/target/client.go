package target

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

// Colly client for interacting with target
type client struct {
	collector *colly.Collector // colly client
	config    clientConfig     // colly client settings (http/TLS timeouts)
	parser    clientParser     // identifies fields to be scraped and parsed from target endpoints
}

// Client configuration
type clientConfig struct {
	urls                map[string]string // holds all url endpoints for a specific target (fixtures and predictions)
	userAgent           string            // set globally or per client
	ignoreRobots        bool              // colly parameter (true, who cares about robots anyway?)
	enableCache         bool              // colly parameter (true, caching is good)
	cacheDir            string            // colly parameter (directory path for cache storage)
	dialTimeout         time.Duration     // colly parameter (request timeout seconds)
	tlsHandShakeTimeout time.Duration     // colly parameter (tls handshake timeout seconds)
}

// Attributes and strings to be scraped and parsed
// on various target endpoints
type clientParser struct {
	login       map[string]string // string identifiers for target login page
	fixtures    map[string]string // string identifiers for target fixture attributes
	results     map[string]string // string identifiers for target fixture results
	predictions map[string]string // string identifiers for target prediction query arguments
}

// Initialise colly client with clientConfig parameters
func (c *client) init(cookieJar http.CookieJar) error {

	// Creates and configures a colly instance with caching
	if c.config.enableCache {
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
			colly.CacheDir(c.config.cacheDir),
		)
	} else {
		// Creates and configures a colly instance without caching
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
		)
	}
	// Sets transport and TLS timeouts,
	// these may need to be relaxed in brubots config.yaml
	// if frequent timeouts occur.
	c.collector.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Second * c.config.dialTimeout,
		}).DialContext,
		TLSHandshakeTimeout: time.Second * c.config.tlsHandShakeTimeout,
	})

	c.collector.IgnoreRobotsTxt = c.config.ignoreRobots

	// Authentication to target is handled within auth.go
	// confirm at a minimum the cookie jar housing authentication
	// token is not empty before setting (we have missed auth in that case).
	// Note, its still possible the cookie has expired and if so, retry auth.
	if cookieJar != nil {
		c.collector.SetCookieJar(cookieJar)
	} else {
		return errors.New("An error occurred setting cookieJar on client, cookieJar is empty")
	}

	return nil

}
