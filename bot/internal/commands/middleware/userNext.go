package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
)

//TODO: add mutex

var UserNext tools.AsyncMap[int64, bot.Command]

func Init() {
	UserNext = tools.AsyncMap[int64, bot.Command]{
		Mutex: sync.Mutex{},
		List:  make(map[int64]bot.Command),
	}
}

func GetUsersNext() bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		userID := update.Message.From.ID
		UserNext.Mutex.Lock()
		if command, ok := UserNext.List[userID]; ok {
			UserNext.Mutex.Unlock()
			defer func() {
				UserNext.Mutex.Lock()
				UserNext.List[userID] = nil
				UserNext.Mutex.Unlock()
			}()
			return command(ctx, api, update)
		} else {
			UserNext.Mutex.Unlock()
			return nil
		}
	}
}
