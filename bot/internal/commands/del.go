package commands

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

func DelInit(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.FromChat().ID

		inlineKeyboard, err := getInlineKeyboard(ctx, db, update.SentFrom().ID, groupID)
		if err != nil {
			if strings.Contains(err.Error(), "no subs") {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Группа не подписана ни на один канал",
					update.Message.TopicID,
				))
				return nil
			} else {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(
					groupID,
					"Не вышло выполнить команду: ошибка при получении списка подписок группы",
					update.Message.TopicID,
				))
				return err
			}
		}

		msg := tgbotapi.NewMessage(
			groupID,
			"Выберите канал, от которого хотите отписаться",
			update.Message.TopicID,
		)
		msg.ReplyMarkup = inlineKeyboard
		_, err = api.Send(msg)
		return err
	}
}

func Del(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.FromChat().ID

		callbackData := strings.Split(update.CallbackData(), "#")
		if callbackData == nil || len(callbackData) != 4 {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return errors.New("incorrect callback data")
		}

		needUserID, err := strconv.ParseInt(callbackData[1], 10, 64)
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return errors.New("incorrect callback data")
		}

		if needUserID != update.SentFrom().ID {
			_, err = api.Send(
				tgbotapi.NewMessage(
					groupID,
					"@"+update.SentFrom().UserName+" команда использована другим пользователем.",
					update.CallbackQuery.Message.TopicID,
				),
			)
			return err
		}

		channelIDStr := callbackData[2]
		channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
		channelName := tools.GetChannelUsername(callbackData[3])

		if channelID == 0 {
			err = tools.ResponseToCallback(
				api,
				update,
				"Команда отменена",
			)
			return err
		}

		isSubed, err := db.CheckSub(ctx, orm.CheckSubParams{
			Chat:    groupID,
			Channel: channelID,
		})
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return err
		}
		if isSubed == 0 {
			err = tools.ResponseToCallback(
				api,
				update,
				"Группа не подписана на @"+channelName,
			)
			return err
		}

		subCnt, err := db.CountSubsOfChannel(ctx, channelID)
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return err
		}

		if subCnt == 1 {
			fetcher, err := db.GetChannelsFetcher(ctx, channelID)
			if err != nil {
				tools.ResponseToCallbackLogError(
					api,
					update,
					"Не вышло отписаться от канала: internal error.",
				)
				return err
			}
			requestURL := "http://" + fetcher.Ip + ":" + fetcher.Port + "/" + channelIDStr
			err = requests.UnsubscribeFromChannel(requestURL)
			if err != nil {
				tools.ResponseToCallbackLogError(
					api,
					update,
					"Не вышло отписаться от канала: internal error.",
				)
				return err
			}
			err = db.DeleteChannel(ctx, channelID)
			if err != nil {
				tools.ResponseToCallbackLogError(
					api,
					update,
					"Не вышло отписаться от канала: internal error.",
				)
				return err
			}
		}

		err = db.UnSubscribe(ctx, orm.UnSubscribeParams{
			Chat:    groupID,
			Channel: channelID,
		})
		if err != nil {
			tools.ResponseToCallbackLogError(
				api,
				update,
				"Не вышло отписаться от канала: internal error.",
			)
			return err
		}

		err = tools.ResponseToCallback(
			api,
			update,
			"Группа успешно отписалась от @"+channelName,
		)
		return err
	}
}

func getInlineKeyboard(ctx context.Context, db *orm.Queries, userID int64, groupID int64) (tgbotapi.InlineKeyboardMarkup, error) {
	channels, err := db.GetGroupSubs(ctx, groupID)
	if err != nil {
		return tgbotapi.InlineKeyboardMarkup{}, err
	}
	if len(channels) == 0 {
		return tgbotapi.InlineKeyboardMarkup{}, errors.New("no subs for group")
	}

	ids := make([]string, len(channels))
	for i, channel := range channels {
		ids[i] = strconv.Itoa(int(channel.ID))
	}

	textID := strconv.Itoa(int(userID))

	rows := make([][]tgbotapi.InlineKeyboardButton, (len(channels)-1)/2+2)
	rowIndex := 0
	for i := 0; i < len(channels); i += 2 {
		if i == len(channels)-1 {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i].Username,
					"del#"+textID+"#"+ids[i]+"#"+channels[i].Username,
				),
			}
		} else {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i].Username,
					"del#"+textID+"#"+ids[i]+"#"+channels[i].Username,
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i+1].Username,
					"del#"+textID+"#"+ids[i+1]+"#"+channels[i+1].Username,
				),
			}
		}
		rowIndex++
	}
	rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"Отмена",
			"del#"+textID+"#0#",
		),
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}
