package bot

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"sync"
	"time"
)

// key is message which was reposted and value is channels to which it was reposted
type channelReposts struct {
	reposts map[tools.MessageConfig][]tools.RepostedTo
	mutex   sync.Mutex
}

// key is message which was edited, value is username of channel
type channelEdit struct {
	edits map[tools.MessageConfig]string
	mutex sync.Mutex
}

type Bot struct {
	api       *tgbotapi.BotAPI
	db        *orm.Queries
	commands  map[string]tools.Command
	callbacks map[string]tools.Command
	channelReposts
	channelEdit
	timeout       time.Duration
	removeGarbage <-chan time.Time
}

func Init(api *tgbotapi.BotAPI, db *orm.Queries, timeout time.Duration, garbageTimeout time.Duration) *Bot {
	return &Bot{
		api:       api,
		db:        db,
		commands:  make(map[string]tools.Command),
		callbacks: make(map[string]tools.Command),
		channelReposts: channelReposts{
			reposts: make(map[tools.MessageConfig][]tools.RepostedTo),
			mutex:   sync.Mutex{},
		},
		channelEdit: channelEdit{
			edits: make(map[tools.MessageConfig]string),
			mutex: sync.Mutex{},
		},
		removeGarbage: time.NewTicker(garbageTimeout).C,
		timeout:       timeout,
	}
}

func (b *Bot) RegisterCommand(name string, command tools.Command) {
	b.commands[name] = command
}

func (b *Bot) RegisterCallback(name string, callback tools.Command) {
	b.callbacks[name] = callback
}

func (b *Bot) Run() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	updates := b.api.GetUpdatesChan(updateConfig)

	log.Println("Waiting for commands...")

	for {
		select {
		case <-b.removeGarbage:
			go b.removeGarbageData()
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
	if update.FromChat() != nil && update.Message != nil && update.Message.MigrateFromChatID != 0 {
		log.Printf("migrate from: %v, to: %v", update.Message.MigrateFromChatID, update.FromChat().ID)
		return b.db.GroupIDChanged(ctx, orm.GroupIDChangedParams{
			update.Message.MigrateFromChatID,
			update.FromChat().ID,
		})
	}

	if update.SentFrom() == nil {
		return nil
	}

	if isFetcher, err := b.isFromFetcher(update); err != nil {
		return err
	} else if isFetcher {
		return b.handleFromFetcher(ctx, update)
	}

	if update.Message != nil {
		msgCmd := update.Message.Command()

		if msgCmd != "" {
			log.Printf(
				"message: chat id: %v, message id: %v, username: %s, cmd: %s",
				update.FromChat().ID,
				update.Message.MessageID,
				update.SentFrom().UserName,
				msgCmd,
			)
		}

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
			update.FromChat().ID,
			update.CallbackQuery.Data,
		)

		// '#' is a separator when bot receives update with callback query between its name and actual data from it
		sepIndex := strings.Index(update.CallbackQuery.Data, "#")
		callbackCmd := update.CallbackQuery.Data[:sepIndex]
		if cmd, ok := b.callbacks[callbackCmd]; ok {
			return cmd(ctx, b.api, update)
		} else {
			return tools.ResponseToCallback(b.api, update, "Несуществующая комманда")
		}
	}
	return nil
}

func (b *Bot) handleFromFetcher(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("message is not from fetcher: message empty")
	}

	/*if update.Message.ForwardFrom != nil {
		log.Printf("%+v\n%+v\n%v", update.Message.ForwardFrom, update.Message, update.Message.ForwardFromMessageID)
	}*/

	if update.Message.ForwardFromChat == nil && update.Message.ForwardFrom == nil {
		if len(update.Message.Text) > 2 {
			if update.Message.Text[0] == 'r' { // got repost message config (ex: "r cID1 mID2 cID3 username")
				cfg, rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
				if err != nil {
					return err
				}
				b.channelReposts.mutex.Lock()
				if b.channelReposts.reposts[*cfg] == nil {
					b.channelReposts.reposts[*cfg] = make([]tools.RepostedTo, 0)
				}
				b.channelReposts.reposts[*cfg] = append(b.channelReposts.reposts[*cfg], *rep)
				//log.Printf("added to reposts from: %+v to: %+v (%v)", *cfg, *rep, len(b.channelReposts.reposts[*cfg]))
				b.channelReposts.mutex.Unlock()
				return nil
			} else if update.Message.Text[0] == 'e' { // got edit message config (ex: "e cID1 mID2 username")
				cfg, channelName, err := tools.GetValuesFromEditConfig(update.Message.Text[2:])
				if err != nil {
					return err
				}
				b.channelEdit.mutex.Lock()
				b.channelEdit.edits[*cfg] = channelName
				b.channelEdit.mutex.Unlock()
				//log.Printf("setted edit: %+v = %s", *cfg, channelName)
				return nil
			} else {
				return errors.New("message is not from fetcher: incorrect code in the beginning")
			}
		} else {
			return errors.New("message is not from fetcher: not forwarded and text length <= 2")
		}
	}

	var chatID int64
	var err error
	if update.Message.ForwardFromChat != nil {
		chatID, err = tools.GetChannelID(update.Message.ForwardFromChat.ID)
		if err != nil {
			return err
		}
		log.Printf("Channel: ID: %v, Name: %s", chatID, update.Message.ForwardFromChat.UserName)
	} else {
		chatID = update.Message.ForwardFrom.ID
		log.Printf("User: ID: %v, Name: %s", chatID, update.Message.ForwardFrom.UserName)
	}

	messageID := update.Message.ForwardFromMessageID
	msgCfg := tools.MessageConfig{
		ChannelID: chatID,
		MessageID: messageID,
	}

	var groups []int64
	var wasRepostOrEdit bool

	b.channelEdit.mutex.Lock()
	if channelName, ok := b.channelEdit.edits[msgCfg]; ok {
		delete(b.channelEdit.edits, msgCfg)
		b.channelEdit.mutex.Unlock()

		wasRepostOrEdit = true

		groups, err = b.db.GetSubsOfChannel(ctx, chatID)
		if err != nil {
			return err
		}

		for _, group := range groups {
			_, err = b.api.Send(tgbotapi.NewMessage(group, "@"+channelName+" отредактировал сообщение:"))
			if err != nil {
				log.Printf("Error sending edited post from channel %v to group %v: %v", chatID, group, err)
				continue
			}

			_, err = b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
			if err != nil {
				log.Printf("Error sending edited post from channel %v to group %v: %v", chatID, group, err)
			}
		}
	} else {
		b.channelEdit.mutex.Unlock()
	}
	b.channelReposts.mutex.Lock()
	if targets, ok := b.channelReposts.reposts[msgCfg]; ok {
		delete(b.channelReposts.reposts, msgCfg)
		b.channelReposts.mutex.Unlock()

		wasRepostOrEdit = true

		for _, target := range targets {
			groups, err = b.db.GetSubsOfChannel(ctx, target.ChannelID)
			if err != nil {
				log.Printf(
					"Could not repost message from channel %v, to channel (%v, %s): %v",
					chatID,
					target.ChannelID,
					target.ChannelName,
					err,
				)
				continue
			}

			for _, group := range groups {
				_, err = b.api.Send(tgbotapi.NewMessage(group, "@"+target.ChannelName+" переслал сообщение:"))
				if err != nil {
					log.Printf("Error sending repost from channel %v to group %v: %v", chatID, group, err)
					continue
				}

				_, err = b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
				if err != nil {
					log.Printf("Error sending repost from channel %v to group %v: %v", chatID, group, err)
				}
			}
		}
	} else {
		b.channelReposts.mutex.Unlock()
	}

	if wasRepostOrEdit || update.Message.ForwardFrom != nil {
		return nil
	}

	groups, err = b.db.GetSubsOfChannel(ctx, chatID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		b.tryUpdateChannelName(ctx, chatID, update.Message.ForwardFromChat.UserName)
		_, err := b.api.Send(tgbotapi.NewForward(group, update.Message.Chat.ID, update.Message.MessageID))
		if err != nil {
			log.Printf("Error sending forward from channel %v to group %v: %v", chatID, group, err)
		}
	}
	return nil
}

func (b *Bot) isFromFetcher(update tgbotapi.Update) (bool, error) {
	isFetcher, err := b.db.CheckFetcher(context.Background(), update.SentFrom().ID)
	if err != nil {
		return false, err
	}
	return isFetcher == 1, nil
}

func (b *Bot) tryUpdateChannelName(ctx context.Context, channelID int64, channelName string) {
	if err := b.db.ChangeChannelUsername(ctx, orm.ChangeChannelUsernameParams{
		ID:       channelID,
		Username: channelName,
	}); err != nil {
		log.Printf("Error changing channel (%v) name to %v: %v", channelID, channelName, err)
	}
}

func (b *Bot) removeGarbageData() {
	b.channelReposts.mutex.Lock()
	b.channelReposts.reposts = make(map[tools.MessageConfig][]tools.RepostedTo)
	b.channelReposts.mutex.Unlock()

	b.channelEdit.mutex.Lock()
	b.channelEdit.edits = make(map[tools.MessageConfig]string)
	b.channelEdit.mutex.Unlock()
}
