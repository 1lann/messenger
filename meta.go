package messenger

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type meta struct {
	req      int64
	revision string
	dtsg     string
	ttstamp  string
}

func (s *Session) populateMeta() error {
	req, _ := http.NewRequest(http.MethodGet, facebookURL, nil)
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.Contains(data,
		[]byte(`<h2 id="security_check_header">Security check</h2>`)) {
		return ErrLoginCheckpoint
	}

	s.meta.dtsg, err = searchBetween(data, "name=\"fb_dtsg\" value=\"", '"')
	if err != nil {
		return err
	}

	s.meta.revision, err = searchBetween(data, "revision\":", ',')
	if err != nil {
		return err
	}

	s.meta.ttstamp = ""
	byteDtsg := []byte(s.meta.dtsg)
	for _, b := range byteDtsg {
		s.meta.ttstamp = s.meta.ttstamp + strconv.Itoa(int(b))
	}
	s.meta.ttstamp = s.meta.ttstamp + "2"

	return nil
}

func searchBetween(data []byte, head string, tail byte) (string, error) {
	i := bytes.Index(data, []byte(head))
	if i < 0 {
		return "", ParseError{"head could not be found"}
	}

	pos := i + len(head)

	var result []byte
	for {
		if data[pos] == tail {
			return strings.TrimSpace(string(result)), nil
		}
		result = append(result, data[pos])
		pos++
	}
}

func (s *Session) addFormMeta(form url.Values) url.Values {
	form.Set("__user", s.userID)
	form.Set("__req", strconv.FormatInt(s.meta.req, 36))
	s.meta.req++
	form.Set("__rev", s.meta.revision)
	form.Set("__a", "1")
	form.Set("__af", "h0")
	form.Set("__be", "-1")
	form.Set("fb_dtsg", s.meta.dtsg)
	form.Set("ttstamp", s.meta.ttstamp)
	return form
}
