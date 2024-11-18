package commands

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Start redirects to /help
func Start(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
	return Help(ctx, api, update)
}
