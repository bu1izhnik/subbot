package requests

import (
	"encoding/json"
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
	if err != nil || res.StatusCode != http.StatusOK {
		return nil, err
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
	if err != nil || res.StatusCode != http.StatusCreated {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	channel := ChannelData{}
	if err := decoder.Decode(&channel); err != nil {
		return nil, err
	}
	return &channel, nil
}
