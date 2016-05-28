package messenger

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Session represents a Facebook session.
type Session struct {
	client   *http.Client
	userID   string
	clientID string

	l    listener
	meta meta
}

// NewSession creates a new Facebook session.
func NewSession() *Session {
	jar, _ := cookiejar.New(nil)

	return &Session{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errNoRedirects
			},
			Jar:     jar,
			Timeout: time.Second * 70,
		},
		meta: meta{
			req: 1,
		},
	}
}
