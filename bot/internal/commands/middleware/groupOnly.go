package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GroupOnly(next bot.Command) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if !update.FromChat().IsGroup() {
			_, err := api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Эту комманду можно использовать только в группах"))
			return err
		}
		err := next(ctx, api, update)
		return err
	}
}
