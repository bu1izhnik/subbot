package bot

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

type Command func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error

type Bot struct {
	api      *tgbotapi.BotAPI
	db       *orm.Queries
	timeout  time.Duration
	commands map[string]Command
}

func Init(api *tgbotapi.BotAPI, db *orm.Queries, timeout time.Duration) *Bot {
	return &Bot{
		api:      api,
		db:       db,
		timeout:  timeout,
		commands: make(map[string]Command),
	}
}

func (b *Bot) RegisterCommand(name string, command Command) {
	b.commands[name] = command
}

func (b *Bot) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	log.Println("Waiting for commands...")

	for {
		select {
		case update := <-updates:
			err := b.HandleUpdate(context.Background(), update)
			if err != nil {
				log.Printf("Error handling update: %v", err)
			}
		}
	}
}

func (b *Bot) HandleUpdate(context context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	log.Printf("chat id: %v, message id: %v", update.Message.Chat.ID, update.Message.MessageID)

	//TODO: Handle forwarded messages from fetchers
	//TODO: Add /help

	msgCmd := update.Message.Command()
	cmd, ok := b.commands[msgCmd]
	if !ok {
		_, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Несуществующая комманда"))
		return err
	}
	return cmd(context, b.api, update)
}
