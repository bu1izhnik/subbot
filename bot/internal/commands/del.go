package commands

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func DelNext(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		middleware.UserNext.Mutex.Lock()
		//middleware.UserNext.List[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(del(db)))
		middleware.UserNext.List[update.SentFrom().ID] =
			middleware.CallbackOnly(
				middleware.GroupOnly(
					middleware.AdminOnly(del(db))))
		middleware.UserNext.Mutex.Unlock()

		groupID := update.FromChat().ID

		inlineKeyboard, err := getInlineKeyboard(ctx, db, groupID)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: ошибка при получении списка подписок группы."))
			return err
		}

		msg := tgbotapi.NewMessage(groupID, "Выберите канал, от которого хотите отписаться.")
		msg.ReplyMarkup = inlineKeyboard
		_, err = api.Send(msg)
		return err
	}
}

func del(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.FromChat().ID

		callbackData := update.CallbackData()
		callbackData, found := strings.CutPrefix(callbackData, "del#")
		if !found {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return errors.New("incorrect callback data")
		}

		channelName := tools.GetChannelUsername(callbackData)

		fetcher, err := tools.GetFetcher(ctx, db, tools.GetMostFullFetcher)
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return err
		}

		requestURL := "http://" + fetcher.Ip + ":" + fetcher.Port + "/" + channelName
		channel, err := requests.ResolveChannelName(requestURL)
		if err != nil {
			if strings.Contains(err.Error(), "channel name") {
				tools.ResponseToCallbackLogError(
					api,
					update,
					"Не вышло отписаться от канала: неверное имя канала.",
				)
			} else {
				tools.ResponseToCallbackLogError(
					api,
					update,
					"Не вышло отписаться от канала: internal error.",
				)
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

		/*isSubed, err := db.CheckSubscription(ctx, orm.CheckSubscriptionParams{
			Chat:    groupID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}
		if isSubed == 0 {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(groupID, "Группа и так не подписана на @"+channel.Username))
			return errors.New("incorrect request: group is not subscribed on this channel")
		}*/

		err = db.UnSubscribe(ctx, orm.UnSubscribeParams{
			Chat:    groupID,
			Channel: channel.ChannelID,
		})
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return err
		}

		err = tools.ResponseToCallback(api, update, "Группа успешно отписалась от @"+channelName)
		/*msg := tgbotapi.NewMessage(groupID, "Группа успешно отписалась от @"+channelName+".")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
		_, err = api.Send(msg)*/
		return err
	}
}

func getInlineKeyboard(ctx context.Context, db *orm.Queries, groupID int64) (tgbotapi.InlineKeyboardMarkup, error) {
	channels, err := db.GetUsernamesOfGroupSubs(ctx, groupID)
	if err != nil {
		return tgbotapi.InlineKeyboardMarkup{}, err
	}

	rows := make([][]tgbotapi.InlineKeyboardButton, (len(channels)-1)/2+1)
	rowIndex := 0
	for i := 0; i < len(channels); i += 2 {
		if i == len(channels)-1 {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i],
					"del#"+channels[i],
				),
			}
		} else {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i],
					"del#"+channels[i],
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i+1],
					"del#"+channels[i+1],
				),
			}
		}
		rowIndex++
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}
