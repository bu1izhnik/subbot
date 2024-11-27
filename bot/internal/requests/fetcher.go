package requests

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ChannelData struct {
	Username   string `json:"username"`
	ChannelID  int64  `json:"channel_id"`
	AccessHash int64  `json:"access_hash"`
}

func ResolveChannelName(reqURL string) (*ChannelData, error) {
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusBadRequest {
		return nil, errors.New("could not resolve channel name")
	} else if res.StatusCode == http.StatusForbidden {
		return nil, errors.New("channel has forbidden forwards")
	}

	decoder := json.NewDecoder(res.Body)
	channel := ChannelData{}
	err = decoder.Decode(&channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func SubscribeToChannel(reqURL string) (*ChannelData, error) {
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusCreated {
		return nil, errors.New("could not subscribe to channel")
	}

	decoder := json.NewDecoder(res.Body)
	channel := ChannelData{}
	if err := decoder.Decode(&channel); err != nil {
		return nil, err
	}
	return &channel, nil
}

func UnsubscribeFromChannel(reqURL string) error {
	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("could not unsubscribe from channel")
	}
	return nil
}
