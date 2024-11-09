package middleware

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CallbackOnly(next tools.Command) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if update.CallbackQuery != nil {
			return next(ctx, api, update)
		}

		return errors.New("not a callback")
	}
}
