package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

var UserNext tools.AsyncMap[int64, tools.Command]

func Init() {
	UserNext = tools.AsyncMap[int64, tools.Command]{
		Mutex: sync.Mutex{},
		List:  make(map[int64]tools.Command),
	}
}

func GetUsersNext() tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		userID := update.SentFrom().ID

		UserNext.Mutex.Lock()
		command, ok := UserNext.List[userID]
		UserNext.Mutex.Unlock()

		if ok {
			err := command(ctx, api, update)
			UserNext.Mutex.Lock()
			delete(UserNext.List, userID)
			UserNext.Mutex.Unlock()
			return err
		} else {
			return nil
		}
	}
}
