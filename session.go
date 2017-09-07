package messenger

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	"time"
)

// Session represents a Facebook session.
type Session struct {
	client       *http.Client
	userID       string
	clientID     string
	requestMutex *sync.RWMutex

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
		requestMutex: new(sync.RWMutex),
		meta: meta{
			req: 1,
		},
	}
}

func (s *Session) doRequest(req *http.Request) (resp *http.Response, err error) {
	s.requestMutex.RLock()
	if os.Getenv("MDEBUG") == "true" {
		log.Println("performing " + req.Method + " request to " + req.URL.String())
	}

	resp, err = s.client.Do(req)

	if os.Getenv("MDEBUG") == "true" {
		log.Println("response code:", resp.Status)
	}
	s.requestMutex.RUnlock()
	return
}
