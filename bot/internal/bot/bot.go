package bot

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"time"
)

type Command func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error

type Bot struct {
	api      *tgbotapi.BotAPI
	db       *orm.Queries
	timeout  time.Duration
	commands map[string]Command
	callback map[string]Command
}

func Init(api *tgbotapi.BotAPI, db *orm.Queries, timeout time.Duration) *Bot {
	return &Bot{
		api:      api,
		db:       db,
		timeout:  timeout,
		commands: make(map[string]Command),
		callback: make(map[string]Command),
	}
}

func (b *Bot) RegisterCommand(name string, command Command) {
	b.commands[name] = command
}

func (b *Bot) RegisterCallback(name string, callback Command) {
	// '#' is a separator when bot receives update with callback query between its name and actual data from it
	b.callback[name+"#"] = callback
}

func (b *Bot) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	log.Println("Waiting for commands...")

	for {
		select {
		case update := <-updates:
			go func(update tgbotapi.Update) {
				err := b.handleUpdate(context.Background(), update)
				if err != nil {
					log.Printf("Error handling update: %v", err)
				}
			}(update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) error {
	if update.Message != nil {
		log.Printf(
			"message: chat id: %v, message id: %v",
			update.Message.Chat.ID,
			update.Message.MessageID,
		)

		if isFetcher, err := b.isFromFetcher(update); err != nil {
			return err
		} else if isFetcher {
			return b.forwardFromFetcher(ctx, update)
		}

		msgCmd := update.Message.Command()

		if cmd, ok := b.commands[msgCmd]; ok {
			return cmd(ctx, b.api, update)
		} else {
			_, err := b.api.Send(
				tgbotapi.NewMessage(update.Message.Chat.ID,
					"Несуществующая комманда",
				))
			return err
		}
	} else if update.CallbackQuery != nil {
		log.Printf(
			"callback query: chat id: %v, query: %v",
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Data,
		)

		sepIndex := strings.Index(update.CallbackQuery.Data, "#")
		callbackCmd := update.CallbackQuery.Data[:sepIndex]
		if cmd, ok := b.commands[callbackCmd]; ok {
			return cmd(ctx, b.api, update)
		} else {
			_, err := b.api.Send(tgbotapi.NewMessage(
				update.CallbackQuery.Message.Chat.ID,
				"Несуществующая комманда",
			))
			return err
		}
	}
	return nil
}

func (b *Bot) forwardFromFetcher(ctx context.Context, update tgbotapi.Update) error {
	if update.Message.ForwardFromChat == nil {
		return errors.New("message is not a forward from channel")
	}

	channelID, err := tools.GetChannelID(update.Message.ForwardFromChat.ID)
	if err != nil {
		return err
	}
	log.Printf("Channel ID: %v", channelID)

	groups, err := b.db.GetSubsOfChannel(ctx, channelID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		b.tryUpdateChannelName(ctx, channelID, update.Message.ForwardFromChat.UserName)
		_, err := b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
		if err != nil {
			log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
		}
	}
	return nil
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
