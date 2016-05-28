package messenger

import (
	"bytes"
	"encoding/gob"
	"net/http"
)

type sessionDump struct {
	FBCookies   []*http.Cookie
	EdgeCookies []*http.Cookie
}

// DumpSession dumps the session (i.e. cookies) and returns it as a []byte.
// Note that if you restore the session, you may not need to login, but you
// must reconnect to chat.
func (s *Session) DumpSession() ([]byte, error) {
	fbCookies := s.client.Jar.Cookies(fbURL)
	edgeCookies := s.client.Jar.Cookies(edgeURL)

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(sessionDump{
		FBCookies:   fbCookies,
		EdgeCookies: edgeCookies,
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RestoreSession restores the session (i.e. cookies) stored as a []byte
// back into the session. Note that you may not need to login again, but you
// must reconnect to chat.
func (s *Session) RestoreSession(data []byte) error {
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	restoredSession := sessionDump{}
	err := dec.Decode(&restoredSession)
	if err != nil {
		return err
	}

	s.client.Jar.SetCookies(fbURL, restoredSession.FBCookies)
	s.client.Jar.SetCookies(edgeURL, restoredSession.EdgeCookies)

	return nil
}

func init() {
	gob.Register(sessionDump{})
}
