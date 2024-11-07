package tools

import (
	"errors"
	"strconv"
	"strings"
)

func GetMessageIDFromMessage(msg string) (int, error) {
	messageIDIndex := strings.Index(msg, " ID:")
	fromIDIndex := strings.Index(msg, " FromID:")
	if messageIDIndex == -1 || fromIDIndex == -1 {
		return -1, errors.New("unexpected message format")
	}
	messageIDStr := msg[messageIDIndex+4 : fromIDIndex]
	messageID, err := strconv.Atoi(messageIDStr)
	return messageID, err
}

func GetChannelIDFromChannel(channel string) (int64, error) {
	channelIDIndex := strings.Index(channel, " ID:")
	accessHashIndex := strings.Index(channel, " AccessHash:")
	if channelIDIndex == -1 || accessHashIndex == -1 {
		return -1, errors.New("unexpected message format")
	}
	channelIDStr := channel[channelIDIndex+4 : accessHashIndex]
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	return channelID, err
}

func GetAccessHashFromChannel(channel string) (int64, error) {
	accessHashIndex := strings.Index(channel, " AccessHash:")
	titleIndex := strings.Index(channel, " Title:")
	if accessHashIndex == -1 || titleIndex == -1 {
		return -1, errors.New("unexpected message format")
	}
	accessHashStr := channel[accessHashIndex+12 : titleIndex]
	accessHash, err := strconv.ParseInt(accessHashStr, 10, 64)
	return accessHash, err
}
