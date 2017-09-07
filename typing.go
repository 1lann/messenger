package messenger

import (
	"net/http"
	"net/url"
	"strings"
)

// SetTypingIndicator sets the typing indicator seen by members of the
// thread.
func (s *Session) SetTypingIndicator(thread Thread, typing bool) error {
	if true {
		return nil
	}

	form := make(url.Values)

	form.Set("source", "mercury-chat")
	form.Set("thread", thread.ThreadID)

	if typing {
		form.Set("typ", "1")
	} else {
		form.Set("typ", "0")
	}

	form.Set("to", "")
	if !thread.IsGroup {
		form.Set("to", thread.ThreadID)
	}

	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, typingURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return err
	}

	resp.Body.Close()

	return nil
}
