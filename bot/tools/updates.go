package tools

import (
	"errors"
	"strconv"
	"strings"
)

// GetValuesFromEditConfig String must contain just data separated by spaces
func GetValuesFromEditConfig(cfg string) (*EditConfig, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 2 {
		return nil, errors.New("invalid edit config")
	}
	edit := &EditConfig{}
	var err error
	edit.ChannelName = data[0]
	edit.Cnt, err = strconv.Atoi(data[1])
	if err != nil {
		return nil, err
	}
	return edit, nil
}

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

// GetValuesFromNotForwardConfig String must contain just data separated by spaces
func GetValuesFromNotForwardConfig(cfg string) (*NotForwardConfig, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, errors.New("invalid wierd config")
	}
	w := &NotForwardConfig{}
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

// GetValuesFromLinkConfig String must contain just data separated by spaces
func GetValuesFromLinkConfig(cfg string) (*ChannelInfo, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 2 {
		return nil, errors.New("invalid link config")
	}
	info := &ChannelInfo{}
	var err error
	info.ID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, err
	}
	info.Name = data[1]
	return info, nil
}
