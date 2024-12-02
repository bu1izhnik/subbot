package commands

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func SubInit(db *redis.Client) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		err := db.Set(
			ctx,
			fmt.Sprintf(
				"next:%d:%d",
				update.SentFrom().ID,
				update.Message.TopicID,
			),
			"sub",
			time.Hour,
		).Err()
		_, err = api.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Отправьте ссылку или юзернейм канала, на который надо подписаться",
			update.Message.TopicID,
		))
		return err
	}
}

func Sub(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.Message.Chat.ID
		channelName := tools.GetChannelUsername(update.Message.Text)

		channels, err := db.GetGroupSubs(ctx, groupID)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(
				groupID,
				"Не вышло подписаться на канал: internal error",
				update.Message.TopicID,
			))
			return err
		}

		if len(channels) >= config.SubLimit {
			_, err = api.Send(tgbotapi.NewMessage(
				groupID,
				fmt.Sprintf("Не вышло подписаться на канал: достигнут максимум в %v подписок", config.SubLimit),
				update.Message.TopicID,
			))
			return err
		}

		fetcher, err := tools.GetFetcher(ctx, db, tools.LeastFull)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(
				groupID,
				"Не вышло подписаться на канал: internal error",
				update.Message.TopicID,
			))
			return err
		}

		requestURL := "http://" + fetcher.Ip + ":" + fetcher.Port + "/" + channelName
		channelCheck, err := requests.ResolveChannelName(requestURL)
		if err != nil {
			if strings.Contains(err.Error(), "name") {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло подписаться на канал: неверное имя канала",
					update.Message.TopicID,
				))
				return err
			} else if strings.Contains(err.Error(), "forwards") {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло подписаться на канал: из канала нельзя пересылать сообщения",
					update.Message.TopicID,
				))
				return err
			} else {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло подписаться на канал: internal error",
					update.Message.TopicID,
				))
				return err
			}
		}

		cnt, err := db.ChannelAlreadyStored(ctx, channelCheck.ChannelID)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(
				groupID,
				"Не вышло подписаться на канал: internal error",
				update.Message.TopicID,
			))
			return err
		}

		for _, alreadySubChannel := range channels {
			if alreadySubChannel.ID == channelCheck.ChannelID {
				_, err = api.Send(tgbotapi.NewMessage(
					groupID,
					"Группа уже подписана на @"+channelName,
					update.Message.TopicID,
				))
				return err
			}
		}

		channel := channelCheck
		// If channel isn't parsed by some fetcher already sub some fetcher to it
		if cnt == 0 {
			channel, err = requests.SubscribeToChannel(requestURL)
			if err != nil {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло подписаться на канал: internal error",
					update.Message.TopicID,
				))
				return err
			}

			_, err = db.AddChannel(ctx, orm.AddChannelParams{
				ID:       channel.ChannelID,
				Hash:     channel.AccessHash,
				Username: channel.Username,
				StoredAt: fetcher.ID,
			})
			if err != nil {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло подписаться на канал: internal error",
					update.Message.TopicID,
				))
				return err
			}
		}

		_, err = db.Subscribe(ctx, orm.SubscribeParams{
			Chat:    update.Message.Chat.ID,
			Channel: channel.ChannelID,
			Thread:  int64(update.Message.TopicID),
		})
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(
				groupID,
				"Не вышло подписаться на канал: internal error",
				update.Message.TopicID,
			))
			return err
		}

		_, err = api.Send(tgbotapi.NewMessage(
			groupID,
			"Группа успешно подписанна на @"+channelName,
			update.Message.TopicID,
		))
		return err
	}
}
