package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func GroupOnly(next tools.Command) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if !(update.FromChat().IsGroup() || update.FromChat().IsSuperGroup()) {
			log.Printf("Chat type isn't group: %s", update.FromChat().Type)
			_, err := api.Send(tgbotapi.NewMessage(
				update.FromChat().ID,
				"Эту комманду можно использовать только в группах",
				update.Message.TopicID,
			))
			return err
		}

		err := next(ctx, api, update)
		return err
	}
}
