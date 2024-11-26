package commands

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

var forms = [3]string{
	"канал",
	"канала",
	"каналов",
}

func List(db *orm.Queries) tools.Command {
	return func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
		subs, err := db.GetGroupSubs(ctx, update.Message.Chat.ID)
		if err != nil {
			tools.SendErrorMessage(api, tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при получении подписок группы"))
			return err
		}

		builder := strings.Builder{}
		if len(subs) == 0 {
			builder.WriteString("Эта группа не подписана ни на один канал")
		} else {
			builder.WriteString(fmt.Sprintf("Группа подписана на %v %s:", len(subs), formOfWord(len(subs))))
			for _, sub := range subs {
				builder.WriteString("\n@")
				builder.WriteString(sub.Username)
			}
		}
		_, err = api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, builder.String()))
		return err
	}
}

func formOfWord(number int) string {
	cases := []int{2, 0, 1, 1, 1, 2}
	var currentCase int
	if number%100 > 4 && number%100 < 20 {
		currentCase = 2
	} else if number%10 < 5 {
		currentCase = cases[number%10]
	} else {
		currentCase = cases[5]
	}
	return forms[currentCase]
}
