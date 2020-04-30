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

type auth struct {
	url            string
	parameters     map[string]string
	passwordEncode bool
	payload        string
	method         string
	userAgent      string
	errorMsg       string
	timeout        time.Duration
	headers        map[string]string
	cookieJar      http.CookieJar
}

func (a *auth) createPayoad() {

	if a.passwordEncode {
		a.parameters["password"] = url.QueryEscape(a.parameters["password"])
	}

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

func (a *auth) authenticate(timeout time.Duration) error {

	a.createPayoad()

	a.cookieJar, _ = cookiejar.New(nil)
	httpClient := http.Client{Jar: a.cookieJar, Timeout: time.Second * a.timeout}

	req, err := http.NewRequest(a.method, a.url, strings.NewReader(a.payload))
	if err != nil {
		return err
	}

	for headerType, headerVal := range a.headers {
		req.Header.Add(headerType, headerVal)
	}

	// global set separately
	req.Header.Set("User-Agent", a.userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("authentication 200 response failure")
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(respBodyBytes), a.errorMsg) {
		return errors.New("authentication incorrect credentials")
	}

	return nil

}
