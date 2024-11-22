package bot

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/internal/commands/middleware"
	"github.com/BulizhnikGames/subbot/bot/internal/config"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
	"sync"
	"time"
)

type Bot struct {
	api *tgbotapi.BotAPI
	db  *orm.Queries

	commands  map[string]tools.Command
	callbacks map[string]tools.Command

	// key is message which was reposted and value is channels to which it was reposted
	channelReposts tools.AsyncMap[tools.MessageConfig, []tools.RepostedTo]
	// key is message which was edited and value is username of channel
	channelEdit tools.AsyncMap[tools.MessageConfig, string]
	// key is id of user and value is his message count per last interval of checks and his ban time of exists
	usersLimits tools.AsyncMap[int64, *tools.RateLimitConfig]
	// key is group id of multimedia and value is messages to forward
	multiMediaQueue tools.AsyncMap[int64, *tools.MultiMediaConfig]

	config.RateLimitConfig

	checkRateLimits       <-chan time.Time
	removeGarbage         <-chan time.Time
	maxMultiMediaWaitTime time.Duration
	timeout               time.Duration
}

func Init(api *tgbotapi.BotAPI,
	db *orm.Queries,
	timeout time.Duration,
	garbageTimeout time.Duration,
	rateLimitCfg config.RateLimitConfig,
	multiMediaWaitTime time.Duration) *Bot {
	return &Bot{
		api: api,
		db:  db,

		commands:  make(map[string]tools.Command),
		callbacks: make(map[string]tools.Command),

		channelReposts: tools.AsyncMap[tools.MessageConfig, []tools.RepostedTo]{
			List:  make(map[tools.MessageConfig][]tools.RepostedTo),
			Mutex: sync.Mutex{},
		},
		channelEdit: tools.AsyncMap[tools.MessageConfig, string]{
			List:  make(map[tools.MessageConfig]string),
			Mutex: sync.Mutex{},
		},
		usersLimits: tools.AsyncMap[int64, *tools.RateLimitConfig]{
			List:  make(map[int64]*tools.RateLimitConfig),
			Mutex: sync.Mutex{},
		},
		multiMediaQueue: tools.AsyncMap[int64, *tools.MultiMediaConfig]{
			List:  make(map[int64]*tools.MultiMediaConfig),
			Mutex: sync.Mutex{},
		},

		RateLimitConfig: rateLimitCfg,

		checkRateLimits:       time.NewTicker(time.Duration(rateLimitCfg.RateLimitCheckInterval) * time.Second).C,
		removeGarbage:         time.NewTicker(garbageTimeout).C,
		maxMultiMediaWaitTime: multiMediaWaitTime,
		timeout:               timeout,
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
		case <-b.checkRateLimits:
			go b.checkForRateLimits()
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
			if msgCmd != "" {
				return middleware.CheckRateLimit(b, cmd)(ctx, b.api, update)
			} else {
				// no need in middleware because it's already implemented in GetUserNext func
				return cmd(ctx, b.api, update)
			}
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
			return middleware.CheckRateLimit(b, cmd)(ctx, b.api, update)
		} else {
			return tools.ResponseToCallback(b.api, update, "Несуществующая комманда")
		}
	}
	return nil
}
