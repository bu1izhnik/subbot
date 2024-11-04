package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AdminOnly(next bot.Command) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		admins, err := api.GetChatAdministrators(
			tgbotapi.ChatAdministratorsConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: update.FromChat().ID,
				},
			},
		)
		if err != nil {
			return err
		}

		for _, admin := range admins {
			if admin.User.ID == update.SentFrom().ID {
				return next(ctx, api, update)
			}
		}

		_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Эту комманду может использовать только админ"))
		return err
	}
}
