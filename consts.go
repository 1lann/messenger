// Package messenger allows you to interact with Facebook chat/Messenger using
// an unofficial API ported from https://github.com/Schmavery/facebook-chat-api.
package messenger

import (
	"errors"
	"net/http"
	"net/url"
)

const (
	facebookURL    = "https://www.facebook.com/"
	facebookOrigin = "https://www.facebook.com"
	loginURL       = "https://www.facebook.com/login.php?login_attempt=1&lwv=110"
	chatURL        = "https://0-edge-chat.facebook.com/pull?"
	threadSyncURL  = "https://www.facebook.com/ajax/mercury/thread_sync.php"
	reconnectURL   = "https://www.facebook.com/ajax/presence/reconnect.php?reason=6"
	readStatusURL  = "https://www.facebook.com/ajax/mercury/change_read_status.php"
	sendMessageURL = "https://www.facebook.com/ajax/mercury/send_messages.php"
	typingURL      = "https://www.facebook.com/ajax/messaging/typ.php"
	syncURL        = "https://www.facebook.com/notifications/sync/?"
	profileURL     = "https://www.facebook.com/chat/user_info/"
	allProfileURL  = "https://www.facebook.com/chat/user_info_all"
	userAgent      = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
	formURLEncoded = "application/x-www-form-urlencoded"
	loggedOutError = 1357001
)

var (
	errNoRedirects = errors.New("no redirects")
	fbURL, _       = url.Parse("https://www.facebook.com")
	edgeURL, _     = url.Parse("https://0-edge-chat.facebook.com")
)

func defaultHeader() http.Header {
	header := make(http.Header)
	header.Set("User-Agent", userAgent)
	header.Set("Origin", facebookOrigin)
	header.Set("Referer", facebookURL)
	return header
}
