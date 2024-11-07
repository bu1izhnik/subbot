package commands

import (
	"context"
	"encoding/json"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/bot"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
)

func delNext(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		middleware.UserNext.Mutex.Lock()
		middleware.UserNext.List[update.Message.From.ID] = middleware.GroupOnly(middleware.AdminOnly(del(db)))
		middleware.UserNext.Mutex.Unlock()
		_, err := api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Отправьте ссылку или юзернейм канала, от которого хотите отписаться."))
		return err
	}
}

func del(db *orm.Queries) bot.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		groupID := update.Message.Chat.ID
		channelName := tools.GetChannelUsername(update.Message.Text)

		fetcherAdr, err := db.GetMostFullFetcher(context.Background())
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
			return err
		}

		requestURL := "http://" + fetcherAdr.Ip + ":" + fetcherAdr.Port + "/" + channelName
		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
			return err
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil || res.StatusCode != http.StatusOK {
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
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
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло отписаться от канала: internal error."))
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
			tools.SendWithErrorLogging(api, tgbotapi.NewMessage(groupID, "Не вышло подписаться на канал: internal error."))
			return err
		}

		_, err = api.Send(tgbotapi.NewMessage(groupID, "Группа успешно отписалась от @"+channelName+"."))
		return err
	}
}
