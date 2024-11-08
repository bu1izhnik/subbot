package bot

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
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
			err := b.handleUpdate(context.Background(), update)
			if err != nil {
				log.Printf("Error handling update: %v", err)
			}
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return nil
	}

	log.Printf("chat id: %v, message id: %v", update.Message.Chat.ID, update.Message.MessageID)

	if isFetcher, err := b.isFromFetcher(update); err != nil {
		return err
	} else if isFetcher {
		go b.forwardFromFetcher(ctx, update)
		return nil
	}

	msgCmd := update.Message.Command()
	cmd, ok := b.commands[msgCmd]
	if !ok {
		_, err := b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Несуществующая комманда"))
		return err
	}
	go func(ctx context.Context, update tgbotapi.Update) {
		err := cmd(ctx, b.api, update)
		if err != nil {
			log.Printf("Error executing user's command: %v", err)
		}
	}(ctx, update)
	return nil
}

func (b *Bot) forwardFromFetcher(ctx context.Context, update tgbotapi.Update) {
	if update.Message.ForwardFromChat == nil {
		log.Printf("Message is not a forward from channel")
		return
	}

	channelID, err := tools.GetChannelID(update.Message.ForwardFromChat.ID)
	if err != nil {
		log.Printf("Error getting channel ID: %v", err)
		return
	}

	log.Printf("Channel ID: %v", channelID)

	groups, err := b.db.GetSubsOfChannel(ctx, channelID)
	if err != nil {
		log.Printf("Error getting subs of channel: %v", err)
		return
	}

	for _, group := range groups {
		b.tryUpdateChannelName(ctx, channelID, update.Message.ForwardFromChat.UserName)
		_, err := b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
		if err != nil {
			log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
		}
	}
}

func (b *Bot) isFromFetcher(update tgbotapi.Update) (bool, error) {
	if update.Message.Text == "" {
		return false, nil
	}
	isFetcher, err := b.db.CheckFetcher(context.Background(), update.Message.From.ID)
	if err != nil {
		return false, err
	}
	if isFetcher == 1 {
		return true, nil
	}
	return false, nil
}

func (b *Bot) tryUpdateChannelName(ctx context.Context, channelID int64, channelName string) {
	if err := b.db.ChangeChannelUsername(ctx, orm.ChangeChannelUsernameParams{
		ID:       channelID,
		Username: channelName,
	}); err != nil {
		log.Printf("Error changing channel (%v) name to %v: %v", channelID, channelName, err)
	}
}
