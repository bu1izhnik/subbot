package tools

import (
	"errors"
	"strconv"
	"strings"
)

// String must contain just data separated by spaces
func GetValuesFromEditConfig(cfg string) (*MessageConfig, string, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, "", errors.New("invalid edit config")
	}
	msgCfg := &MessageConfig{}
	var err error
	msgCfg.ChannelID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, "", err
	}
	msgCfg.MessageID, err = strconv.Atoi(data[1])
	if err != nil {
		return nil, "", err
	}
	return msgCfg, data[2], nil
}

// String must contain just data separated by spaces
func GetValuesFromRepostConfig(cfg string) (*MessageConfig, *RepostedTo, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 4 {
		return nil, nil, errors.New("invalid repost config")
	}
	msgCfg := &MessageConfig{}
	repTo := &RepostedTo{}
	var err error
	msgCfg.ChannelID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	msgCfg.MessageID, err = strconv.Atoi(data[1])
	if err != nil {
		return nil, nil, err
	}
	repTo.ChannelID, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	repTo.ChannelName = data[3]
	return msgCfg, repTo, nil
}
