package messenger

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func generateClientID() string {
	data := make([]byte, 4)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(data)
}

type accessibilityStruct struct {
	Sr    int64 `json:"sr"`
	SrTs  int64 `json:"sr-ts"`
	Jk    int64 `json:"jk"`
	JkTs  int64 `json:"jk-ts"`
	Kb    int64 `json:"kb"`
	KbTs  int64 `json:"kb-ts"`
	Hcm   int64 `json:"hcm"`
	HcmTs int64 `json:"hcm-ts"`
}

func generateAccessibilityCookie() string {
	now := time.Now().UnixNano() / 1000000

	access := accessibilityStruct{
		Sr:    0,
		SrTs:  now,
		Jk:    0,
		JkTs:  now,
		Kb:    0,
		KbTs:  now,
		Hcm:   0,
		HcmTs: now,
	}

	res, err := json.Marshal(access)
	if err != nil {
		panic(err)
	}

	return url.QueryEscape(string(res))
}

// ConnectToChat connects the session to chat after you've successfully
// logged in.
func (s *Session) ConnectToChat() error {
	err := s.populateMeta()
	if err != nil {
		return err
	}

	s.l.form = s.newPullForm()

	err = s.requestReconnect()
	if err != nil {
		return err
	}

	err = s.connectToStage1()
	if err != nil {
		return err
	}

	err = s.connectToStage2()
	if err != nil {
		return err
	}

	err = s.connectToStage3()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) requestReconnect() error {
	req, _ := http.NewRequest(http.MethodGet, reconnectURL, nil)
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	resp.Body.Close()
	return nil
}

func (s *Session) connectToStage1() error {
	req, err := s.createStage1Request()
	if err != nil {
		return err
	}

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respInfo, err := parseResponse(resp.Body)
	if err != nil {
		return err
	}

	if respInfo.Type != "lb" {
		return ParseError{"non t: \"lb\" response from chat server"}
	}

	s.l.form.stickyPool = respInfo.Sticky.Pool
	s.l.form.stickyToken = respInfo.Sticky.Token

	return nil
}

func (s *Session) createStage1Request() (*http.Request, error) {
	cookies := s.client.Jar.Cookies(fbURL)
	for _, cookie := range cookies {
		if cookie.Name == "c_user" {
			s.userID = cookie.Value
			break
		}
	}

	if s.userID == "" {
		return nil, ParseError{"missing required c_user user ID"}
	}

	s.clientID = generateClientID()
	presence := s.generatePresence()

	cookies = append(cookies, []*http.Cookie{
		{
			Name:   "presence",
			Value:  presence,
			Domain: ".facebook.com",
			Secure: true,
		},
		{
			Name:   "locale",
			Value:  "en_US",
			Domain: ".facebook.com",
			Secure: true,
		},
		{
			Name:   "a11y",
			Value:  generateAccessibilityCookie(),
			Domain: ".facebook.com",
			Secure: true,
		},
	}...)

	s.client.Jar.SetCookies(fbURL, cookies)

	form := s.newPullForm()

	req, _ := http.NewRequest(http.MethodGet,
		chatURL+form.form().Encode(), nil)
	req.Header = defaultHeader()

	return req, nil
}

func (s *Session) connectToStage2() error {
	req, _ := http.NewRequest(http.MethodGet,
		chatURL+s.l.form.form().Encode(), nil)
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}

func (s *Session) connectToStage3() error {
	form := make(url.Values)
	form.Set("client", "mercury")
	form.Set("folders[0]", "inbox")
	form.Set("last_action_timestamp", "0")
	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, threadSyncURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}
