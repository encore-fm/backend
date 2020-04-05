package systest

import (
	"fmt"
	"net/http"
)

var (
	BackendBaseUrl = "http://127.0.0.1:8080"
)

func UserJoin(username string, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/join/%v", BackendBaseUrl, username, sessionID)

	return http.Post(endpointUrl, "", nil)
}

func UserPing(username, secret, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/ping", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)
	return client.Do(req)
}

func UserList(username string, secret string, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/list", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

func UserSuggestSong(username string, secret string, sessionID string, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/suggest/%v", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

func UserListSongs(username string, secret string, sessionID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/listSongs", BackendBaseUrl, username)

	client := &http.Client{}
	req, err := http.NewRequest("GET", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

func UserUpvote(username string, secret string, sessionID string, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/vote/%v/up", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

func UserDownvote(username string, secret string, sessionID string, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/vote/%v/down", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("POST", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}

func AdminRemoveSong(username string, secret string, sessionID string, songID string) (resp *http.Response, err error) {
	endpointUrl := fmt.Sprintf("%v/users/%v/removeSong/%v", BackendBaseUrl, username, songID)

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", endpointUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Session", sessionID)
	req.Header.Set("Authorization", secret)

	return client.Do(req)
}
