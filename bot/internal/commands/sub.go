package commands

import (
	"context"
	"encoding/json"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
)

func SubNext(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		middleware.UserNext[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(sub(db)))
		_, err := api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Отправьте ссылку или юзернейм канала, на который надо подписаться."))
		return err
	}
}

func sub(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		channels, err := db.ListGroupSubs(ctx, update.Message.Chat.ID)
		if err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		if len(channels) >= config.SUB_LIMIT {
			_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: достигнут максимум в 5 подписок."))
			return err
		}

		channelName := tools.GetChannelUsername(update.Message.Text)

		fetcherAdr, err := db.GetLeastFullFetcher(context.Background())
		if err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		requestURL := "http://" + fetcherAdr.Ip + ":" + fetcherAdr.Port + "/" + channelName
		req, err := http.NewRequest(http.MethodPost, requestURL, nil)
		if err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != http.StatusCreated {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		type channelData struct {
			Username   string `json:"username"`
			ChannelID  int64  `json:"channel_id"`
			AccessHash int64  `json:"access_hash"`
		}
		decoder := json.NewDecoder(res.Body)
		channel := channelData{}
		if err := decoder.Decode(&channel); err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		for _, ID := range channels {
			if ID == channel.ChannelID {
				_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Группа уже подписана на @"+channelName+"."))
				return err
			}
		}

		_, err = db.Subscribe(ctx, orm.SubscribeParams{
			Chat:    update.Message.Chat.ID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		_, err = db.AddChannel(ctx, orm.AddChannelParams{
			ID:       channel.ChannelID,
			Hash:     channel.AccessHash,
			Username: channel.Username,
		})
		if err != nil {
			api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Группа успешно подписанна на канал @"+channelName+"."))
		return err
	}
}
