package messenger

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Errors that are returned by Login.
var (
	ErrLoginError      = errors.New("messenger: incorrect login credentials")
	ErrLoginCheckpoint = errors.New("messenger: login checkpoint")
)

var jsCookiePattern = regexp.MustCompile("\\[\"(_js_[^\"]+)\",\"([^\"]+)\",")

func (s *Session) createLoginRequest(email, password string) (*http.Request, error) {
	req, _ := http.NewRequest(http.MethodGet, facebookURL, nil)
	req.Header = defaultHeader()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	form := make(url.Values)

	doc.Find("#login_form input").Each(func(i int, s *goquery.Selection) {
		name, found := s.Attr("name")
		if !found {
			return
		}

		value, found := s.Attr("value")
		if !found {
			return
		}

		form.Set(name, value)
	})

	cookies := s.client.Jar.Cookies(fbURL)

	matches := jsCookiePattern.FindAllStringSubmatch(string(data), -1)
	for _, match := range matches {
		cookies = append(cookies, &http.Cookie{
			Name:   match[1],
			Value:  strings.Replace(match[2], "\\/", "/", -1),
			Domain: "facebook.com",
		})
	}

	s.client.Jar.SetCookies(fbURL, cookies)

	form.Set("email", email)
	form.Set("pass", password)
	form.Set("default_persistent", "1")
	form.Set("lgnjs", strconv.FormatInt(time.Now().Unix(), 10))
	_, offset := time.Now().Zone()
	form.Set("timezone", strconv.Itoa(-offset/60))
	form.Set("lgndim", "eyJ3IjoxNDQwLCJoIjo5MDAsImF3IjoxNDQwLCJhaCI6OTAwLCJjIjoyNH0=")
	form.Set("next", "https://www.facebook.com/")

	loginReq, _ := http.NewRequest(http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	loginReq.Header = defaultHeader()
	loginReq.Header.Set("Content-Type", formURLEncoded)

	return loginReq, nil
}

// Login logs the session in to a Facebook account.
func (s *Session) Login(email, password string) error {
	req, err := s.createLoginRequest(email, password)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err == nil {
		resp.Body.Close()
		return ErrLoginError
	}

	urlErr, ok := err.(*url.Error)
	if !ok || urlErr.Err != errNoRedirects {
		return err
	}

	err = handleLoginRedirect(resp)
	if err != nil {
		return err
	}

	return nil
}

func handleLoginRedirect(resp *http.Response) error {
	redirURL, err := resp.Location()
	if err != nil {
		return err
	}

	if strings.Contains(redirURL.String(), "https://www.facebook.com/checkpoint") {
		return ErrLoginCheckpoint
	}

	if strings.Contains(redirURL.String(), "https://www.facebook.com/login.php?") {
		return ErrLoginError
	}

	if redirURL.String() == "https://www.facebook.com/" || redirURL.String() == "https://www.facebook.com" {
		return nil
	}

	return ParseError{"unexpected redirect to " + redirURL.String()}
}
