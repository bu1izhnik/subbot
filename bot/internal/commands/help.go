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
				"/list - посмотреть на какие каналы подписана группа\n"+
				"/del - отписать группу от канала\n"+
				"/help - базовая информация о боте\n"+
				"\n"+
				"Дополнительная информация:\n"+
				"Бот может временно игнорировать пользователей, которые излишне спамят командами.\n"+
				"Бот имеет открытый исходный код: https://github.com/BulizhnikGames/subbot",
			config.SubLimit,
		),
	))
	return err
}
