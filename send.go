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
		"client":                                            []string{"mercury"},
		"message_batch[0][action_type]":                     []string{"ma-type:user-generated-message"},
		"message_batch[0][author]":                          []string{"fbid:" + s.userID},
		"message_batch[0][timestamp]":                       []string{strconv.FormatInt(time.Now().UnixNano()/1e6, 10)},
		"message_batch[0][timestamp_absolute]":              []string{"Today"},
		"message_batch[0][timestamp_relative]":              []string{time.Now().Format("15:04")},
		"message_batch[0][timestamp_time_passed]":           []string{"0"},
		"message_batch[0][is_unread]":                       []string{"false"},
		"message_batch[0][is_cleared]":                      []string{"false"},
		"message_batch[0][is_forward]":                      []string{"false"},
		"message_batch[0][is_filtered_content]":             []string{"false"},
		"message_batch[0][is_filtered_content_bh]":          []string{"false"},
		"message_batch[0][is_filtered_content_account]":     []string{"false"},
		"message_batch[0][is_filtered_content_quasar]":      []string{"false"},
		"message_batch[0][is_filtered_content_invalid_app]": []string{"false"},
		"message_batch[0][is_spoof_warning]":                []string{"false"},
		"message_batch[0][source]":                          []string{"source:chat:web"},
		"message_batch[0][source_tags][0]":                  []string{"source:chat"},
		"message_batch[0][body]":                            []string{msg.Body},
		"message_batch[0][html_body]":                       []string{"false"},
		"message_batch[0][ui_push_phase]":                   []string{"V3"},
		"message_batch[0][status]":                          []string{"0"},
		"message_batch[0][offline_threading_id]":            []string{msg.offlineThreadID},
		"message_batch[0][message_id]":                      []string{msg.offlineThreadID},
		"message_batch[0][threading_id]":                    []string{s.generateThreadID()},
		"message_batch[0][ephemeral_ttl_mode]:":             []string{"0"},
		"message_batch[0][manual_retry_cnt]":                []string{"0"},
		"message_batch[0][has_attachment]":                  []string{hasAttachment},
		"message_batch[0][signatureID]":                     []string{generateSignatureID()},
	}

	if msg.Thread.IsGroup {
		form.Set("message_batch[0][thread_fbid]", msg.Thread.ThreadID)
	} else {
		form.Set("message_batch[0][specific_to_list][0]", "fbid:"+
			msg.Thread.ThreadID)
		form.Set("message_batch[0][specific_to_list][1]", "fbid:"+s.userID)
		form.Set("message_batch[0][other_user_fbid]", msg.Thread.ThreadID)
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
