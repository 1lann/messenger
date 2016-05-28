package messenger

import (
	"net/http"
	"net/url"
	"strings"
)

// UserProfile represents the profile information for a user.
type UserProfile struct {
	UserID    string `json:"id"`
	Name      string `json:"name"`
	FirstName string `json:"firstName"`
	Vanity    string `json:"vanity"`
	IsFriend  bool   `json:"is_friend"`
}

type singleUserResponse struct {
	Payload struct {
		Profiles map[string]UserProfile `json:"profiles"`
	} `json:"payload"`
	Error int `json:"error"`
}

// UserProfileInfo returns the user's profile given their ID.
func (s *Session) UserProfileInfo(userID string) (UserProfile, error) {
	form := make(url.Values)
	form.Set("ids[0]", userID)
	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, profileURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.client.Do(req)
	if err != nil {
		return UserProfile{}, err
	}

	defer resp.Body.Close()

	var singleUserResp singleUserResponse
	err = unmarshalPullData(resp.Body, &singleUserResp)
	if err != nil {
		return UserProfile{}, err
	}

	if singleUserResp.Error == loggedOutError {
		return UserProfile{}, ErrLoggedOut
	} else if singleUserResp.Error > 0 {
		return UserProfile{}, ErrUnknown
	}

	profile, found := singleUserResp.Payload.Profiles[userID]
	if !found {
		return UserProfile{}, ParseError{"could not find userID in response"}
	}

	return profile, nil
}

type allUsersResponse struct {
	Payload map[string]UserProfile `json:"payload"`
	Error   int                    `json:"error"`
}

// AllUserProfileInfo returns all the users' profiles in the session's friend
// list as a map indexed by the user's ID.
func (s *Session) AllUserProfileInfo() (map[string]UserProfile, error) {
	form := make(url.Values)
	form.Set("viewer", s.userID)
	form = s.addFormMeta(form)

	req, _ := http.NewRequest(http.MethodPost, allProfileURL,
		strings.NewReader(form.Encode()))
	req.Header = defaultHeader()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var allUsersResp allUsersResponse
	err = unmarshalPullData(resp.Body, &allUsersResp)
	if err != nil {
		return nil, err
	}

	if allUsersResp.Error == loggedOutError {
		return nil, ErrLoggedOut
	} else if allUsersResp.Error > 0 {
		return nil, ErrUnknown
	}

	return allUsersResp.Payload, nil
}
