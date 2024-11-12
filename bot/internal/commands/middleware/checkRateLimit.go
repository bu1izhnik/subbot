package middleware

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type checker interface {
	IncreaseMsgCountForUser(userID int64) bool
}

// CheckRateLimit Auto-implemented in user next middleware
func CheckRateLimit(bot checker, next tools.Command) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		if can := bot.IncreaseMsgCountForUser(update.SentFrom().ID); can {
			return next(ctx, api, update)
		} else {
			return nil
		}
	}
}
