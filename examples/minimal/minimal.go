// This is an example of a Facebook messenger chat bot which repeats
// messages it receives, being as minimal as possible without persisting
// sessions, and thus whose configuration is not recommended. Refer to
// the `repeat` example instead.

// The email and password used is taken from the environment variables
// FBEMAIL and FBPASS.

package main

import (
	"log"
	"os"

	"github.com/1lann/messenger"
)

func main() {
	s := messenger.NewSession()

	err := s.Login(os.Getenv("FBEMAIL"), os.Getenv("FBPASS"))
	if err != nil {
		panic(err)
	}

	err = s.ConnectToChat()
	if err != nil {
		panic(err)
	}

	s.OnMessage(func(msg *messenger.Message) {
		log.Println("Received \"" + msg.Body + "\" from " + msg.FromUserID)

		resp := s.NewMessageWithThread(msg.Thread)
		resp.Body = "You said: " + msg.Body

		_, err := s.SendMessage(resp)
		if err != nil {
			log.Println("Failed to send message:", err)
		}
	})

	log.Println("Waiting for messages...")
	s.Listen()
}
