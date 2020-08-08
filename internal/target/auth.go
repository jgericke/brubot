package target

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// all the important target auth stuff
type auth struct {
	url            string            // URL string for target authentication
	parameters     map[string]string // Target credentials strng
	passwordEncode bool              // Specifies whether to URL encode credential string
	payload        string            // Holds the auth string built by createPayload()
	method         string            // http method for auth endpoint
	userAgent      string            // user agent string set in request header
	errorMsg       string            // HTML body response string to establish failure
	timeout        time.Duration     // Auth http client timeout seconds
	headers        map[string]string // Headers map to set on auth query
	cookieJar      http.CookieJar    // Returned on successful auth for use by colly
}

// Authenticate builds and sends auth string to target and populates
// a cookiejar to be passed to colly on successful auth.
func (t *Target) Authenticate() error {

	// Call to authenticate method, results in population of auth token
	// within cookiejar
	if err := t.Auth.authenticate(t.Auth.timeout); err != nil {
		return err
	}
	// Initialises client with all client specific parameters, passing
	// auth cookie jar for authenticating subsequent queries.
	if err := t.Client.init(t.Auth.cookieJar); err != nil {
		return err
	}

	return nil

}

// builds auth query, url encoding where required
func (a *auth) createPayoad() {

	if a.passwordEncode {
		a.parameters["password"] = url.QueryEscape(a.parameters["password"])
	}

	// creates a string for authentication from auth.parameters map
	idx := 0
	for param, val := range a.parameters {
		if idx == 0 {
			a.payload += fmt.Sprintf("%s=%s", param, val)
		} else {
			a.payload += fmt.Sprintf("&%s=%s", param, val)
		}
		idx++
	}

}

// Query to target authentication url, using method to submit auth payload
// Checks response to establish success of authentication attemp
func (a *auth) authenticate(timeout time.Duration) error {

	a.createPayoad()

	// create a cookieJar to be passed to colly client
	a.cookieJar, _ = cookiejar.New(nil)
	// std http client setup
	httpClient := http.Client{Jar: a.cookieJar, Timeout: time.Second * a.timeout}

	req, err := http.NewRequest(a.method, a.url, strings.NewReader(a.payload))
	if err != nil {
		return err
	}

	for headerType, headerVal := range a.headers {
		req.Header.Add(headerType, headerVal)
	}

	// global headers are set separately
	req.Header.Set("User-Agent", a.userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// was a 200 response received from auth query
	if resp.StatusCode != http.StatusOK {
		return errors.New("An invalid response code was received when authenticating to target")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse response body for auth failure message and react accordingly
	if strings.Contains(string(respBodyBytes), a.errorMsg) {
		return errors.New("Invalid credentials when authenticating to target")
	}

	return nil

}
