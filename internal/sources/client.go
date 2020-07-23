/*
   This is pretty much a reproduction of target/client.go with *minor* tweaks,
   i.e. no need to auth or mess around with cookies.

   Given that the main driver for retrieving predictions off of source endpoints
   is Colly there is a lot of overlap between sources and targets.

   That said, each source endpoint is unique in terms of parsing requirements,
   while additionally source and target packages may be split into separate services
   at some point. Maybe.
*/

package sources

import (
	"net"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

// Colly client setup
// Created per-source endpoint which might be expensive but given
// each source endpoint will have it's own timeouts and parsing
// properties, this cannot be avoided :/
type client struct {
	collector *colly.Collector
	config    clientConfig
	parser    clientParser
}

// Client setup for each source endpoint
// These are explained better within target.client.clientConfig
type clientConfig struct {
	urls                map[string]string
	userAgent           string
	ignoreRobots        bool
	enableCache         bool
	cacheDir            string
	dialTimeout         time.Duration
	tlsHandShakeTimeout time.Duration
}

// olds attributes for parsing when fetching prediction values from each source endpoint
// predictions:		structured as "attribute name (i.e. on_html): string_to_match"
type clientParser struct {
	predictions map[string]string // string identifiers for target prediction query arguments
}

// Initialise colly client with clientConfig parameters
func (c *client) init() {

	if c.config.enableCache {
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
			colly.CacheDir(c.config.cacheDir),
		)
	} else {
		c.collector = colly.NewCollector(
			colly.UserAgent(c.config.userAgent),
		)
	}

	c.collector.WithTransport(&http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Second * c.config.dialTimeout,
		}).DialContext,
		TLSHandshakeTimeout: time.Second * c.config.tlsHandShakeTimeout,
	})

	c.collector.IgnoreRobotsTxt = c.config.ignoreRobots
}
