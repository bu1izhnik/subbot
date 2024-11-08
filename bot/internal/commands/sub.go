package commands

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func SubNext(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		middleware.UserNext.Mutex.Lock()
		middleware.UserNext.List[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(sub(db)))
		middleware.UserNext.Mutex.Unlock()
		_, err := api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Отправьте ссылку или юзернейм канала, на который надо подписаться."))
		return err
	}
}

func sub(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.Message.Chat.ID
		channelName := tools.GetChannelUsername(update.Message.Text)

		channels, err := db.ListGroupSubs(ctx, groupID)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		if len(channels) >= config.SUB_LIMIT {
			_, err = api.Send(tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: достигнут максимум в 5 подписок."))
			return err
		}

		//TODO: firstly check if channel is already stored at some parser

		var requestURL string
		fetcherAdr, err := db.GetLeastFullFetcher(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				rndFetcherAdr, err := db.GetRandomFetcher(ctx)
				if err != nil {
					tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
					return err
				}
				requestURL = "http://" + rndFetcherAdr.Ip + ":" + rndFetcherAdr.Port + "/" + channelName
			} else {
				tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
				return err
			}
		} else {
			requestURL = "http://" + fetcherAdr.Ip + ":" + fetcherAdr.Port + "/" + channelName
		}

		channelCheck, err := requests.ResolveChannelName(requestURL)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		cnt, err := db.ChannelAlreadyStored(ctx, channelCheck.ChannelID)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		for _, ID := range channels {
			if ID == channelCheck.ChannelID {
				_, err = api.Send(tgbotapi.NewMessage(groupID, "Группа уже подписана на @"+channelName+"."))
				return err
			}
		}

		channel := channelCheck
		if cnt == 0 {
			channel, err = requests.SubscribeToChannel(requestURL)
			if err != nil {
				tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
				return err
			}

			_, err = db.AddChannel(ctx, orm.AddChannelParams{
				ID:       channel.ChannelID,
				Hash:     channel.AccessHash,
				Username: channel.Username,
				StoredAt: fetcherAdr.ID,
			})
			if err != nil {
				tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
				return err
			}
		}

		_, err = db.Subscribe(ctx, orm.SubscribeParams{
			Chat:    update.Message.Chat.ID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		_, err = api.Send(tgbotapi.NewMessage(groupID, "Группа успешно подписанна на канал @"+channelName+"."))
		return err
	}
}
