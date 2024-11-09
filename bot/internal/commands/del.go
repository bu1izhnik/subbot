package commands

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func DelNext(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		middleware.UserNext.Mutex.Lock()
		middleware.UserNext.List[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(del(db)))
		middleware.UserNext.Mutex.Unlock()

		groupID := update.Message.Chat.ID
		inlineKeyboard, err := getInlineKeyboard(ctx, db, groupID)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Отправьте ссылку или юзернейм канала, от которого хотите отписаться."))
			return err
		}

		msg := tgbotapi.NewMessage(groupID, "Отправьте ссылку или юзернейм канала, от которого хотите отписаться.")
		inlineKeyboard.Selective = true
		msg.ReplyMarkup = inlineKeyboard
		_, err = api.Send(msg)
		return err
	}
}

func del(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.Message.Chat.ID
		channelName := tools.GetChannelUsername(update.Message.Text)

		fetcher, err := tools.GetFetcher(ctx, db, tools.GetMostFullFetcher)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
			return err
		}

		requestURL := "http://" + fetcher.Ip + ":" + fetcher.Port + "/" + channelName
		channel, err := requests.ResolveChannelName(requestURL)
		if err != nil {
			if strings.Contains(err.Error(), "channel name") {
				tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: неверное имя канала."))
			} else {
				tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
			}
			return err
		}

		err = db.ChangeChannelUsernameAndHash(ctx, orm.ChangeChannelUsernameAndHashParams{
			ID:       channel.ChannelID,
			Username: channel.Username,
			Hash:     channel.AccessHash,
		})
		if err != nil {
			log.Printf("Error updating channel's username and hash: %v", err)
		}

		isSubed, err := db.CheckSubscription(ctx, orm.CheckSubscriptionParams{
			Chat:    groupID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}
		if isSubed == 0 {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Группа и так не подписана на @"+channel.Username))
			return errors.New("incorrect request: group is not subscribed on this channel")
		}

		err = db.UnSubscribe(ctx, orm.UnSubscribeParams{
			Chat:    groupID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		msg := tgbotapi.NewMessage(groupID, "Группа успешно отписалась от @"+channelName+".")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		_, err = api.Send(msg)
		return err
	}
}

func getInlineKeyboard(ctx context.Context, db *orm.Queries, groupID int64) (tgbotapi.ReplyKeyboardMarkup, error) {
	channels, err := db.GetUsernamesOfGroupSubs(ctx, groupID)
	if err != nil {
		return tgbotapi.ReplyKeyboardMarkup{}, err
	}

	rows := make([][]tgbotapi.KeyboardButton, (len(channels)-1)/2+1)
	rowIndex := 0
	for i := 0; i < len(channels); i += 2 {
		if i == len(channels)-1 {
			rows[rowIndex] = []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("@" + channels[i])}
		} else {
			rows[rowIndex] = []tgbotapi.KeyboardButton{
				tgbotapi.NewKeyboardButton("@" + channels[i]),
				tgbotapi.NewKeyboardButton("@" + channels[i+1]),
			}
		}
		rowIndex++
	}
	return tgbotapi.NewReplyKeyboard(rows...), nil
}
