package messenger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func parseResponse(rd io.Reader) (pullResponse, error) {
	var result pullResponse
	err := unmarshalPullData(rd, &result)
	if err != nil {
		return pullResponse{}, err
	}

	if result.Error == loggedOutError {
		return pullResponse{}, ErrLoggedOut
	} else if result.Error > 0 {
		return pullResponse{}, ErrUnknown
	}

	return result, nil
}

func unmarshalPullData(rd io.Reader, to interface{}) error {
	data, err := ioutil.ReadAll(rd)
	if err != nil {
		return err
	}

	if os.Getenv("MDEBUG") == "true" {
		if len(data) > 10000 {
			log.Println("debug response size:", len(data))
		} else {
			log.Println("debug response: " + string(data))
		}
	}

	startPos := bytes.IndexByte(data, '{')
	if startPos < 0 {
		return ParseError{"could not find start of response"}
	}

	err = json.Unmarshal(data[startPos:], to)
	if err != nil {
		fmt.Println(string(data))
		return err
	}

	return nil
}
