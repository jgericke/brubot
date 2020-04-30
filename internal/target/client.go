package target

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

type client struct {
	collector *colly.Collector
	config    clientConfig
	parser    clientParser
}

type clientConfig struct {
	urls                map[string]string
	userAgent           string
	ignoreRobots        bool
	enableCache         bool
	cacheDir            string
	dialTimeout         time.Duration
	tlsHandShakeTimeout time.Duration
}

type clientParser struct {
	login    map[string]string
	fixtures map[string]string
}

func (c *client) init(cookieJar http.CookieJar) {

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
	c.collector.SetCookieJar(cookieJar)

}

func (c *client) getFixtures(round *Round) error {

	var err error

	c.collector.OnRequest(func(r *colly.Request) {
		log.Println("Retrieving from:", r.URL)
	})

	c.collector.OnHTML(c.parser.fixtures["attr_onhtml"], func(e *colly.HTMLElement) {

		e.ForEach(c.parser.fixtures["attr_fixture"], func(_ int, cl *colly.HTMLElement) {

			leftID, convErr := strconv.Atoi(cl.Attr(c.parser.fixtures["attr_t_leftid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			rightID, convErr := strconv.Atoi(cl.Attr(c.parser.fixtures["attr_t_rightid"]))
			if convErr != nil {
				err = errors.New("failure converting team ID")
				return
			}

			round.Fixtures = append(round.Fixtures, fixture{
				token: cl.Attr(c.parser.fixtures["attr_token"]),
				leftTeam: strings.Split(cl.Attr(c.parser.fixtures["attr_teams"]),
					c.parser.fixtures["attr_teams_delimiter"])[0],
				rightTeam: strings.Split(cl.Attr(c.parser.fixtures["attr_teams"]),
					c.parser.fixtures["attr_teams_delimiter"])[1],
				leftID:  leftID,
				rightID: rightID,
			})

		})

	})

	c.collector.OnHTML(c.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("not authenticated")
		return
	})

	c.collector.OnError(func(r *colly.Response, resError error) {
		log.Printf("error response %+v occurred retrieving from %s message: %s", r, r.Request.URL, resError)
		err = fmt.Errorf("error response %+v occurred retrieving from %s message: %s", r, r.Request.URL, resError)
		return
	})

	c.collector.Visit(fmt.Sprint(c.config.urls["fixtures"], round.id))

	if err != nil {
		return err
	}

	return nil

}
