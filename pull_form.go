package messenger

import (
	"net/url"
	"strconv"
)

type pullForm struct {
	userID      string
	clientID    string
	stickyToken string
	stickyPool  string

	seq              int
	partition        int
	state            string
	cap              int
	messagesReceived int
	idleTime         int
}

func (s *Session) newPullForm() pullForm {
	return pullForm{
		userID:           s.userID,
		clientID:         s.clientID,
		stickyToken:      "",
		stickyPool:       "",
		seq:              0,
		partition:        -2,
		state:            "active",
		cap:              8,
		messagesReceived: 0,
		idleTime:         0,
	}
}

func (p pullForm) encode() string {
	form := url.Values{
		"channel":    []string{"p_" + p.userID},
		"seq":        []string{strconv.Itoa(p.seq)},
		"partition":  []string{strconv.Itoa(p.partition)},
		"clientid":   []string{p.clientID},
		"viewer_uid": []string{p.userID},
		"uid":        []string{p.userID},
		"state":      []string{p.state},
		"idle":       []string{strconv.Itoa(p.idleTime)},
		"cap":        []string{strconv.Itoa(p.cap)},
		"msgs_recv":  []string{strconv.Itoa(p.messagesReceived)},
	}

	if p.stickyPool != "" && p.stickyToken != "" {
		form.Set("sticky_token", p.stickyToken)
		form.Set("sticky_pool", p.stickyPool)
	}

	return form.Encode()
}
