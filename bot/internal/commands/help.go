package commands

import (
	"context"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Help(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error {
	_, err := api.Send(tgbotapi.NewMessage(
		update.FromChat().ID,
		fmt.Sprintf(
			"Данный бот позволяет подписывать группы на телеграм каналы: "+
				"в группу будут пересылаться все новые сообщения из каналов, "+
				"на которые она подписана.\n"+
				"\n"+
				"Доступные команды:\n"+
				"/sub - подписать группу на канал (максимум %v подписок)\n"+
				"Если в группе не один чат, а несколько тем, то бот будет пересылать сообщения из канала в ту тему, где была использована команда\n"+
				"/list - посмотреть на какие каналы подписана группа\n"+
				"/del - отписать группу от канала\n"+
				"/help - базовая информация о боте\n"+
				"\n"+
				"Дополнительная информация:\n"+
				"1. Бот может временно игнорировать пользователей, которые излишне спамят командами.\n"+
				"2. В данный момент бот не подписывается на приватные каналы.\n"+
				"3. Бот имеет открытый исходный код: https://github.com/BulizhnikGames/subbot",
			config.SubLimit,
		),
		update.Message.TopicID,
	))
	return err
}
