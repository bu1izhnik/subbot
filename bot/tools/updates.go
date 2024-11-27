package tools

import (
	"errors"
	"strconv"
	"strings"
)

// GetValuesFromEditConfig String must contain just data separated by spaces
/*func GetValuesFromEditConfig(cfg string) (*MessageConfig, string, error) {
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
}*/

// GetValuesFromRepostConfig String must contain just data separated by spaces
func GetValuesFromRepostConfig(cfg string) (*RepostConfig, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, errors.New("invalid repost config")
	}
	rep := &RepostConfig{}
	var err error
	rep.To.ID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, err
	}
	rep.To.Name = data[1]
	rep.Cnt, err = strconv.Atoi(data[2])
	if err != nil {
		return nil, err
	}
	return rep, nil
}

// GetValuesFromWeirdConfig String must contain just data separated by spaces
func GetValuesFromWeirdConfig(cfg string) (*WeirdConfig, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, errors.New("invalid wierd config")
	}
	w := &WeirdConfig{}
	var err error
	w.Channel.ID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, err
	}
	w.Channel.Name = data[1]
	w.Cnt, err = strconv.Atoi(data[2])
	if err != nil {
		return nil, err
	}
	return w, nil
}
