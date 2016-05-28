package messenger

import (
	"net/http"
	"net/url"
	"strings"
)

// MarkAsRead marks the specified thread as read.
func (s *Session) MarkAsRead(thread Thread) error {
	form := make(url.Values)
	form.Set("ids["+thread.ThreadID+"]", "true")
	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, readStatusURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = parseResponse(resp.Body)
	return err
}
