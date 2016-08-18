package messenger

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Attachment represents an attachment
type Attachment struct {
	Name string
	Data io.Reader
}

// Message represents a message object.
type Message struct {
	FromUserID      string
	Thread          Thread
	Body            string
	Attachments     []Attachment
	MessageID       string
	offlineThreadID string
}

// NewMessageWithThread creates a new message for the given thread.
func (s *Session) NewMessageWithThread(thread Thread) *Message {
	return &Message{
		Thread:          thread,
		offlineThreadID: generateOfflineThreadID(),
	}
}

type sendResponse struct {
	Payload pullMessage `json:"payload"`
	Error   int         `json:"error"`
}

// SendMessage sends the message to the session. Only the Thread, Body and
// Attachments fields are used for sending. The message ID and error is returned.
//
// TODO: Sending does not support attachments yet.
func (s *Session) SendMessage(msg *Message) (string, error) {
	hasAttachment := "false"
	if len(msg.Attachments) > 0 {
		hasAttachment = "true"
	}

	form := url.Values{
		"client":                          []string{"mercury"},
		"action_type":                     []string{"ma-type:user-generated-message"},
		"author":                          []string{"fbid:" + s.userID},
		"timestamp":                       []string{strconv.FormatInt(time.Now().UnixNano()/1e6, 10)},
		"timestamp_absolute":              []string{"Today"},
		"timestamp_relative":              []string{time.Now().Format("15:04")},
		"timestamp_time_passed":           []string{"0"},
		"is_unread":                       []string{"false"},
		"is_cleared":                      []string{"false"},
		"is_forward":                      []string{"false"},
		"is_filtered_content":             []string{"false"},
		"is_filtered_content_bh":          []string{"false"},
		"is_filtered_content_account":     []string{"false"},
		"is_filtered_content_quasar":      []string{"false"},
		"is_filtered_content_invalid_app": []string{"false"},
		"is_spoof_warning":                []string{"false"},
		"source":                          []string{"source:chat:web"},
		"source_tags[0]":                  []string{"source:chat"},
		"body":                            []string{msg.Body},
		"html_body":                       []string{"false"},
		"ui_push_phase":                   []string{"V3"},
		"status":                          []string{"0"},
		"offline_threading_id":            []string{msg.offlineThreadID},
		"message_id":                      []string{msg.offlineThreadID},
		"threading_id":                    []string{s.generateThreadID()},
		"ephemeral_ttl_mode:":             []string{"0"},
		"manual_retry_cnt":                []string{"0"},
		"has_attachment":                  []string{hasAttachment},
		"signatureID":                     []string{generateSignatureID()},
	}

	if msg.Thread.IsGroup {
		form.Set("thread_fbid", msg.Thread.ThreadID)
	} else {
		form.Set("specific_to_list[0]", "fbid:"+
			msg.Thread.ThreadID)
		form.Set("specific_to_list[1]", "fbid:"+s.userID)
		form.Set("other_user_fbid", msg.Thread.ThreadID)
	}

	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, sendMessageURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var respMsg sendResponse
	err = unmarshalPullData(resp.Body, &respMsg)
	if err != nil {
		return "", err
	}

	if respMsg.Error == loggedOutError {
		return "", ErrLoggedOut
	} else if respMsg.Error > 0 {
		return "", ErrUnknown
	}

	if len(respMsg.Payload.Actions) == 0 {
		return "", ParseError{"expected more than 0 actions after sending"}
	}

	messageID := respMsg.Payload.Actions[0].MessageID
	if messageID == "" {
		return "", ParseError{"missing expected message ID"}
	}

	return messageID, nil
}

func generateOfflineThreadID() string {
	random := strconv.FormatInt(largeRandomNumber(), 2)
	if len(random) < 22 {
		random = strings.Repeat("0", 22-len(random)) + random
	} else {
		random = random[:22]
	}

	now := strconv.FormatInt(time.Now().UnixNano()/1e6, 2)
	n, err := strconv.ParseInt(now+random, 2, 64)
	if err != nil {
		// If this happens, it's the end of the world.
		panic(err)
	}

	return strconv.FormatInt(n, 10)
}

func (s *Session) generateThreadID() string {
	now := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	r := strconv.FormatInt(largeRandomNumber(), 10)
	return "<" + now + ":" + r + "-" + s.clientID + "@mail.projektitan.com>"
}

func generateSignatureID() string {
	return strconv.FormatInt(largeRandomNumber()/2-1, 16)
}
