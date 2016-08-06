// This is an example of a Facebook messenger chat bot which repeats
// messages it receives. It also persist sessions by saving sessions
// to a file.

// The email and password used is taken from the environment variables
// FBEMAIL and FBPASS.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/1lann/messenger"
)

const sessionFile = "session_data"

var s *messenger.Session

func main() {
	s = messenger.NewSession()
	login()

	err := s.ConnectToChat()
	if err != nil {
		fmt.Println("Failed to connect to chat:", err)
		return
	}

	fmt.Println("Connected to chat")

	// Save the session file every minute. The use of saved session prevents
	// Facebook from flagging your bot as suspicious for having to re-log so
	// often, and also makes it faster to test your bot as it doesn't need to
	// log in again every time.
	go func() {
		ticker := time.Tick(time.Minute)
		for range ticker {
			saveSession()
		}
	}()

	s.OnMessage(func(msg *messenger.Message) {
		fmt.Println("Received \"" + msg.Body + "\" from " + msg.FromUserID)

		resp := s.NewMessageWithThread(msg.Thread)
		resp.Body = "You said: " + msg.Body

		_, err := s.SendMessage(resp)
		if err != nil {
			fmt.Println("Failed to send message:", err)
		}
	})

	fmt.Println("Waiting for messages...")
	s.Listen()
}

func login() {
	sessionData, err := ioutil.ReadFile(sessionFile)
	if os.IsNotExist(err) {
		fmt.Println("No session file, logging in...")
		err = s.Login(os.Getenv("FBEMAIL"), os.Getenv("FBPASS"))
		if err != nil {
			fmt.Println("Failed to login:", err)
			os.Exit(1)
		}
		return
	}

	err = s.RestoreSession(sessionData)
	if err != nil {
		log.Println("Failed to restore session, logging in...")
		err = s.Login(os.Getenv("FBEMAIL"), os.Getenv("FBPASS"))
		if err != nil {
			fmt.Println("Failed to login:", err)
			os.Exit(1)
		}
		return
	}
}

func saveSession() {
	data, err := s.DumpSession()
	if err != nil {
		fmt.Println("Failed to save session:", err)
		return
	}

	err = ioutil.WriteFile(sessionFile, data, 0644)
	if err != nil {
		fmt.Println("Failed to write session to file:", err)
	}
}
