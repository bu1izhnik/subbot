package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

var NextCommand map[string]tools.Command

func Init() {
	NextCommand = make(map[string]tools.Command)
}

func RegisterCommand(name string, command tools.Command) {
	NextCommand[name] = command
}

func GetUsersNext(bot checker, db *redis.Client) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		userID := update.SentFrom().ID
		topicID := update.Message.TopicID

		key := fmt.Sprintf(
			"next:%d:%d",
			userID,
			topicID,
		)

		textCmd, err := db.Get(ctx, key).Result()
		if err != nil {
			return err
		}

		command, ok := NextCommand[textCmd]

		db.Del(ctx, key)

		if ok {
			err = CheckRateLimit(bot, command)(ctx, api, update)
			return err
		} else {
			return errors.New(fmt.Sprintf(
				"invalid command stored for user %d (topic %d) in redis: %s",
				userID,
				topicID,
				textCmd,
			))
		}
	}
}
