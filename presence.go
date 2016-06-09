package messenger

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/url"
	"strings"
	"time"
)

type presenceStruct struct {
	V     int            `json:"v"`
	Time  int64          `json:"time"`
	User  string         `json:"user"`
	State presenceState  `json:"state"`
	Ch    map[string]int `json:"ch"`
}

type presenceState struct {
	Ut   int         `json:"ut"`
	T2   []int       `json:"t2"`
	Lm2  interface{} `json:"lm2"`
	Uct2 int64       `json:"uct2"`
	Tr   interface{} `json:"tr"`
	Tw   int64       `json:"tw"`
	At   int64       `json:"at"`
}

var presenceDecodeMap = map[string]string{
	"_": "%",
	"A": "%2",
	"B": "000",
	"C": "%7d",
	"D": "%7b%22",
	"E": "%2c%22",
	"F": "%22%3a",
	"G": "%2c%22ut%22%3a1",
	"H": "%2c%22bls%22%3a",
	"I": "%2c%22n%22%3a%22%",
	"J": "%22%3a%7b%22i%22%3a0%7d",
	"K": "%2c%22pt%22%3a0%2c%22vis%22%3a",
	"L": "%2c%22ch%22%3a%7b%22h%22%3a%22",
	"M": "%7b%22v%22%3a2%2c%22time%22%3a1",
	"N": ".channel%22%2c%22sub%22%3a%5b",
	"O": "%2c%22sb%22%3a1%2c%22t%22%3a%5b",
	"P": "%2c%22ud%22%3a100%2c%22lc%22%3a0",
	"Q": "%5d%2c%22f%22%3anull%2c%22uct%22%3a",
	"R": ".channel%22%2c%22sub%22%3a%5b1%5d",
	"S": "%22%2c%22m%22%3a0%7d%2c%7b%22i%22%3a",
	"T": "%2c%22blc%22%3a1%2c%22snd%22%3a1%2c%22ct%22%3a",
	"U": "%2c%22blc%22%3a0%2c%22snd%22%3a1%2c%22ct%22%3a",
	"V": "%2c%22blc%22%3a0%2c%22snd%22%3a0%2c%22ct%22%3a",
	"W": "%2c%22s%22%3a0%2c%22blo%22%3a0%7d%2c%22bl%22%3a%7b%22ac%22%3a",
	"X": "%2c%22ri%22%3a0%7d%2c%22state%22%3a%7b%22p%22%3a0%2c%22ut%22%3a1",
	"Y": "%2c%22pt%22%3a0%2c%22vis%22%3a1%2c%22bls%22%3a0%2c%22blc%22%3a0%2c%22snd%22%3a1%2c%22ct%22%3a",
	"Z": "%2c%22sb%22%3a1%2c%22t%22%3a%5b%5d%2c%22f%22%3anull%2c%22uct%22%3a0%2c%22s%22%3a0%2c%22blo%22%3a0%7d%2c%22bl%22%3a%7b%22ac%22%3a",
}

var presenceEncode [][2]string

func init() {
	for letter := 'Z'; letter >= 'A'; letter-- {
		letterStr := string(letter)
		str := presenceDecodeMap[letterStr]

		presenceEncode = append(presenceEncode, [2]string{letterStr, str})
	}

	presenceEncode = append(presenceEncode, [2]string{"_", "%"})
}

func (s *Session) generatePresence() string {
	now := time.Now()

	presence := presenceStruct{
		V:    3,
		Time: now.Unix(),
		User: s.userID,
		State: presenceState{
			Ut:   0,
			T2:   []int{},
			Lm2:  nil,
			Uct2: now.UnixNano() / 1e6,
			Tr:   nil,
			Tw:   largeRandomNumber(),
			At:   now.UnixNano() / 1e6,
		},
		Ch: map[string]int{
			"p_" + s.userID: 0,
		},
	}

	result, err := json.Marshal(presence)
	if err != nil {
		panic(err)
	}

	return "E" + encodePresence(string(result))
}

func encodePresence(str string) string {
	esc := strings.Replace(strings.ToLower(url.QueryEscape(str)), "_", "%5f", -1)
	for _, line := range presenceEncode {
		esc = strings.Replace(esc, line[1], line[0], -1)
	}
	return esc
}

func decodePresence(str string) (string, error) {
	output := ""
	for i := 0; i < len(str); i++ {
		resolve, found := presenceDecodeMap[string(str[i])]
		if !found {
			output += string(str[i])
			continue
		}

		output += resolve
	}

	return url.QueryUnescape(output)
}

func largeRandomNumber() int64 {
	max := big.NewInt(4294967296)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic(err)
	}

	return n.Int64()
}
