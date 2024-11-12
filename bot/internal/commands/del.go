package commands

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/requests"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func DelInit(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		/*middleware.UserNext.Mutex.Lock()
		//middleware.UserNext.List[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(del(db)))
		middleware.UserNext.List[update.SentFrom().ID] =
			middleware.CallbackOnly(
				middleware.GroupOnly(
					middleware.AdminOnly(del(db))))
		middleware.UserNext.Mutex.Unlock()*/

		groupID := update.FromChat().ID

		inlineKeyboard, err := getInlineKeyboard(ctx, db, update.SentFrom().ID, groupID)
		if err != nil {
			if strings.Contains(err.Error(), "no subs") {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(groupID, "Не вышло выполнить команду: группа не подписана ни на один канал."))
				return nil
			} else {
				tools.SendErrorMessage(api, tgbotapi.NewMessage(groupID, "Не вышло выполнить команду: ошибка при получении списка подписок группы"))
				return err
			}
		}

		msg := tgbotapi.NewMessage(groupID, "Выберите канал, от которого хотите отписаться.")
		msg.ReplyMarkup = inlineKeyboard
		_, err = api.Send(msg)
		return err
	}
}

func Del(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.FromChat().ID

		callbackData := strings.Split(update.CallbackData(), "#")
		if callbackData == nil || len(callbackData) != 3 {
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
			tools.SendErrorMessage(
				api,
				tgbotapi.NewMessage(
					groupID,
					"@"+update.SentFrom().UserName+" команда использована другим пользователем.",
				),
			)
		}

		channelName := tools.GetChannelUsername(callbackData[2])

		if channelName == "" {
			err = tools.ResponseToCallback(
				api,
				update,
				"Команда отменена",
			)
			return err
		}

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

		err = tools.ResponseToCallback(
			api,
			update,
			"Группа успешно отписалась от @"+channelName,
		)
		return err
	}
}

func getInlineKeyboard(ctx context.Context, db *orm.Queries, userID int64, groupID int64) (tgbotapi.InlineKeyboardMarkup, error) {
	channels, err := db.GetUsernamesOfGroupSubs(ctx, groupID)
	if err != nil {
		return tgbotapi.InlineKeyboardMarkup{}, err
	}
	if len(channels) == 0 {
		return tgbotapi.InlineKeyboardMarkup{}, errors.New("no subs for group")
	}

	textID := strconv.Itoa(int(userID))

	rows := make([][]tgbotapi.InlineKeyboardButton, (len(channels)-1)/2+2)
	rowIndex := 0
	for i := 0; i < len(channels); i += 2 {
		if i == len(channels)-1 {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i],
					"del#"+textID+"#"+channels[i],
				),
			}
		} else {
			rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i],
					"del#"+textID+"#"+channels[i],
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"@"+channels[i+1],
					"del#"+textID+"#"+channels[i+1],
				),
			}
		}
		rowIndex++
	}
	rows[rowIndex] = []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			"Отмена",
			"del#"+textID+"#",
		),
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...), nil
}
