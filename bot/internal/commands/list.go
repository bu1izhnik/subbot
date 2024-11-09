package commands

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func List(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		subs, err := db.GetUsernamesOfGroupSubs(ctx, update.Message.Chat.ID)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении подписок группы."))
			return err
		}

		builder := strings.Builder{}
		if len(subs) == 0 {
			builder.WriteString("Эта группа не подписана ни на один канал")
		} else {
			builder.WriteString(fmt.Sprintf("Группа подписана на %v каналов:", len(subs)))
			for _, sub := range subs {
				builder.WriteString("\n@")
				builder.WriteString(sub)
			}
		}
		_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, builder.String()))
		return err
	}
}
