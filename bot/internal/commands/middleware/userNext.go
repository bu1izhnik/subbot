package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var UserNext map[int64]bot.Command

func Init() {
	UserNext = make(map[int64]bot.Command)
}

func GetUsersNext() bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if command, ok := UserNext[update.Message.From.ID]; ok {
			defer func() {
				UserNext[update.Message.From.ID] = nil
			}()
			return command(ctx, api, update)
		} else {
			return nil
		}
	}
}
