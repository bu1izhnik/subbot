package tools

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"sync"
)

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

func GetChannelUsername(username string) string {
	if username[0] == '@' {
		return username[1:]
	} else if strings.HasPrefix(username, "t.me/") {
		return username[5:]
	} else if strings.HasPrefix(username, "https://t.me/") {
		return username[13:]
	}
	return username
}

func SendWithErrorLogging(api *tgbotapi.BotAPI, message tgbotapi.Chattable) {
	_, err := api.Send(message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
