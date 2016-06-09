package messenger

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A subset of possible errors returned by OnError
var (
	ErrLoggedOut = errors.New("messenger: (probably) logged out")
	ErrUnknown   = errors.New("messenger: unknown error from server")
)

type listener struct {
	form pullForm

	lastMessage   time.Time
	activeRequest *http.Request
	lastSync      time.Time
	// TODO: Close functions are hackily thread safe.
	shouldClose bool
	closed      chan bool
	closeMutex  *sync.Mutex

	onMessage func(msg *Message)
	onRead    func(thread Thread, userID string)
	onError   func(err error)

	processedThreadMessages map[string][]string
	processedMutex          *sync.Mutex
}

// ListenError is the type of error that will always be passed to OnError.
// It contains information about the operation that caused the error, and the
// actual underlying error.
type ListenError struct {
	Op  string
	Err error
}

func (l ListenError) Error() string {
	return "listen: " + l.Op + ": " + l.Err.Error()
}

// Listen starts listening for events and messages from Facebook's chat
// servers and blocks.
func (s *Session) Listen() {
	s.l.closeMutex = new(sync.Mutex)

	s.checkListeners()

	s.l.lastMessage = time.Now()
	s.l.lastSync = time.Now()

	go func() {
		for !s.l.shouldClose {
			s.listenRequest()
		}
	}()

	s.l.closeMutex.Lock()
	<-s.l.closed
	s.l.shouldClose = true
	s.l.closeMutex.Unlock()
}

func (s *Session) checkListeners() {
	if s.l.onError == nil {
		s.l.onError = func(err error) { fmt.Println(err) }
	}

	if s.l.onMessage == nil {
		s.l.onMessage = func(msg *Message) {}
	}

	if s.l.onRead == nil {
		s.l.onRead = func(thread Thread, userID string) {}
	}
}

// OnMessage sets the handler for when a message is received.
//
// Receiving attachments isn't supported yet.
func (s *Session) OnMessage(handler func(msg *Message)) {
	s.l.onMessage = handler
}

// OnRead sets the handler for when a message is read.
func (s *Session) OnRead(handler func(thread Thread, userID string)) {
	s.l.onRead = handler
}

// OnError sets the handler for when an error during listening occurs.
func (s *Session) OnError(handler func(err error)) {
	s.l.onError = handler
}

// Close stops and returns all listeners on the session.
func (s *Session) Close() error {
	s.l.closed <- true
	s.l.closeMutex.Lock()
	s.l.closeMutex.Unlock()
	return nil
}

type pullMsgMeta struct {
	Sender    string `json:"actorFbId"`
	ThreadKey struct {
		ThreadID    string `json:"threadFbId"`
		OtherUserID string `json:"otherUserFbId"`
	} `json:"threadKey"`
	MessageID string `json:"messageId"`
	Timestamp string `json:"timestamp"`
}

type pullAction struct {
	ThreadID  string `json:"thread_fbid"`
	Author    string `json:"author"`
	MessageID string `json:"message_id"`
}

type pullMessage struct {
	Type   string `json:"type"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Reader int64  `json:"reader"`
	Delta  struct {
		Class    string      `json:"class"`
		Body     string      `json:"body"`
		Metadata pullMsgMeta `json:"messageMetadata"`
	} `json:"delta"`
	Event      string       `json:"event"`
	Actions    []pullAction `json:"actions"`
	St         int          `json:"st"`
	ThreadID   int64        `json:"thread_fbid"`
	FromMobile bool         `json:"from_mobile"`
	UserID     int64        `json:"realtime_viewer_fbid"`
	Reason     string       `json:"reason"`
}

type pullResponse struct {
	Type   string `json:"t"`
	Sticky struct {
		Token string `json:"sticky"`
		Pool  string `json:"pool"`
	} `json:"lb_info"`
	Seq      int           `json:"seq"`
	Messages []pullMessage `json:"ms"`
	Reason   int           `json:"reason"`
	Error    int           `json:"error"`
}

func (s *Session) listenRequest() {
	idleSeconds := time.Now().Sub(s.l.lastMessage).Seconds()
	s.l.form.idleTime = int(idleSeconds)

	presence := s.generatePresence()
	cookies := s.client.Jar.Cookies(fbURL)
	cookies = append(cookies, &http.Cookie{
		Name:   "presence",
		Value:  presence,
		Domain: ".facebook.com",
	})
	s.client.Jar.SetCookies(fbURL, cookies)

	req, _ := http.NewRequest(http.MethodGet, chatURL+s.l.form.form().Encode(),
		nil)
	req.Header = defaultHeader()

	resp, err := s.doRequest(req)
	if err != nil {
		go s.l.onError(ListenError{"HTTP listen", err})
		time.Sleep(time.Second)
		return
	}

	defer resp.Body.Close()

	respInfo, err := parseResponse(resp.Body)
	if err != nil {
		go s.l.onError(ListenError{"parse listen", err})
		time.Sleep(time.Second)
		return
	}

	s.l.lastMessage = time.Now()
	s.l.form.messagesReceived += len(respInfo.Messages)
	s.l.form.seq = respInfo.Seq

	if respInfo.Type == "refresh" && respInfo.Reason == 110 {
		go s.l.onError(ListenError{"listen response", ErrLoggedOut})
		if !s.l.shouldClose {
			s.l.closed <- true
			s.l.closeMutex.Lock()
			s.l.closeMutex.Unlock()
		}

		return
	}

	if respInfo.Type == "fullReload" {
		if os.Getenv("MDEBUG") == "true" {
			log.Println("debug start full reload")
			s.fullReload()
			log.Println("debug end full reload")
		} else {
			s.fullReload()
		}

		return
	}

	go s.processPull(respInfo)

	time.Sleep(time.Second)
}

func (s *Session) processPull(resp pullResponse) {
	if resp.Type == "lb" {
		s.l.form.stickyToken = resp.Sticky.Token
		s.l.form.stickyPool = resp.Sticky.Pool
	}

	for _, msg := range resp.Messages {
		if msg.Type == "delta" {
			if msg.Delta.Class != "NewMessage" {
				continue
			}

			s.handleDeltaMessage(msg.Delta.Body, msg.Delta.Metadata)
		} else if msg.Type == "messaging" {
			if msg.Event == "read_receipt" {
				thread := Thread{
					ThreadID: strconv.FormatInt(msg.Reader, 10),
					IsGroup:  false,
				}
				if msg.ThreadID != 0 {
					thread.ThreadID = strconv.FormatInt(msg.ThreadID, 10)
					thread.IsGroup = true
				}

				go s.l.onRead(thread, strconv.FormatInt(msg.Reader, 10))
			}
		}
	}
}

func (s *Session) handleDeltaMessage(body string, meta pullMsgMeta) {
	if meta.Sender == s.userID {
		return
	}

	threadID := meta.ThreadKey.ThreadID
	isGroup := true
	if threadID == "" {
		threadID = meta.Sender
		isGroup = false
	}

	msg := &Message{
		FromUserID: meta.Sender,
		Thread: Thread{
			ThreadID: threadID,
			IsGroup:  isGroup,
		},
		Body:      body,
		MessageID: meta.MessageID,
	}

	go s.l.onMessage(msg)
}

func (s *Session) fullReload() {
	func() {
		form := make(url.Values)
		form.Set("lastSync", strconv.FormatInt(s.l.lastSync.Unix(), 10))
		form = s.addFormMeta(form)

		req, _ := http.NewRequest(http.MethodGet, syncURL+form.Encode(), nil)
		req.Header = defaultHeader()

		resp, err := s.doRequest(req)
		if err != nil {
			s.l.onError(ListenError{"reload sync", err})
			return
		}

		s.l.lastSync = time.Now()

		resp.Body.Close()
	}()

	func() {
		form := make(url.Values)
		form.Set("client", "mercury")
		form.Set("folders[0]", "inbox")
		form.Set("last_action_timestamp",
			strconv.FormatInt((time.Now().UnixNano()/1e6)-60, 10))
		form = s.addFormMeta(form)

		req, _ := http.NewRequest(http.MethodPost, threadSyncURL,
			strings.NewReader(form.Encode()))

		resp, err := s.doRequest(req)
		if err != nil {
			s.l.onError(ListenError{"reload thread sync", err})
			return
		}

		resp.Body.Close()
	}()
}
