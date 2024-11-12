package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func AdminOnly(next tools.Command) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if !(update.FromChat().IsGroup() || update.FromChat().IsSuperGroup()) {
			log.Printf("Chat type isn't group: %s", update.FromChat().Type)
			_, err := api.Send(tgbotapi.NewMessage(update.FromChat().ID, "Эту комманду можно использовать только в группах"))
			return err
		}

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

		_, err = api.Send(tgbotapi.NewMessage(update.FromChat().ID, "Эту комманду может использовать только админ"))
		return err
	}
}
