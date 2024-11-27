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
func GetValuesFromRepostConfig(cfg string) (*Repost, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, errors.New("invalid repost config")
	}
	rep := &Repost{}
	var err error
	rep.ChannelID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, err
	}
	rep.ChannelName = data[1]
	rep.Cnt, err = strconv.Atoi(data[2])
	if err != nil {
		return nil, err
	}
	return rep, nil
}
