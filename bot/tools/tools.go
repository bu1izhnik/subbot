package tools

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
	"sync"
)

type MessageConfig struct {
	ChannelID int64
	MessageID int
}

type RepostedTo struct {
	ChannelID   int64
	ChannelName string
}

type FetcherParams struct {
	ID   int64
	Ip   string
	Port string
}

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

type Command func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error

type GetFetcherRequest func(ctx context.Context, db *orm.Queries) (*FetcherParams, error)

// For some reason tgbotapi package adds "-100" to all ID's of channels and supergroups, this "-100" need to be removed
func GetChannelID(id int64) (int64, error) {
	idStr := strconv.FormatInt(id, 10)
	if len(idStr) <= 4 {
		return 0, errors.New("incorrect channel id")
	}
	idStr, _ = strings.CutPrefix(idStr, "-100")
	newID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

func GetChannelUsername(username string) string {
	if username == "" {
		return ""
	}
	if username[0] == '@' {
		return username[1:]
	} else if strings.HasPrefix(username, "t.me/") {
		return username[5:]
	} else if strings.HasPrefix(username, "https://t.me/") {
		return username[13:]
	}
	return username
}

// String must contain just data separated by spaces
func GetValuesFromEditConfig(cfg string) (*MessageConfig, string, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 3 {
		return nil, "", errors.New("invalid edit config")
	}
	msgCfg := &MessageConfig{}
	var err error
	msgCfg.ChannelID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, "", err
	}
	msgCfg.MessageID, err = strconv.Atoi(data[1])
	if err != nil {
		return nil, "", err
	}
	return msgCfg, data[2], nil
}

// String must contain just data separated by spaces
func GetValuesFromRepostConfig(cfg string) (*MessageConfig, *RepostedTo, error) {
	data := strings.Split(cfg, " ")
	if len(data) != 4 {
		return nil, nil, errors.New("invalid repost config")
	}
	msgCfg := &MessageConfig{}
	repTo := &RepostedTo{}
	var err error
	msgCfg.ChannelID, err = strconv.ParseInt(data[0], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	msgCfg.MessageID, err = strconv.Atoi(data[1])
	if err != nil {
		return nil, nil, err
	}
	repTo.ChannelID, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		return nil, nil, err
	}
	repTo.ChannelName = data[3]
	return msgCfg, repTo, nil
}

func SendErrorMessage(api *tgbotapi.BotAPI, message tgbotapi.Chattable) {
	_, err := api.Send(message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func ResponseToCallback(api *tgbotapi.BotAPI, update tgbotapi.Update, newText string) error {
	if update.CallbackQuery == nil {
		return errors.New("no callback query to response with error")
	}

	groupID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID

	_, err := api.Send(tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      groupID,
			MessageID:   messageID,
			ReplyMarkup: nil,
		},
	})
	if err != nil {
		return err
	}

	_, err = api.Send(tgbotapi.NewEditMessageText(
		groupID,
		messageID,
		newText,
	))
	return err
}

func ResponseToCallbackLogError(api *tgbotapi.BotAPI, update tgbotapi.Update, newText string) {
	err := ResponseToCallback(api, update, newText)
	if err != nil {
		log.Printf("Error responding to callback query: %v", err)
	}
}

// Trying to get fetcher with providing func, if fails gets random fetcher, if it also fails returns error
func GetFetcher(ctx context.Context, db *orm.Queries, get GetFetcherRequest) (*FetcherParams, error) {
	fetcher, err := get(ctx, db)
	if err != nil {
		randomFetcher, err := db.GetRandomFetcher(ctx)
		if err != nil {
			return nil, err
		}
		return &FetcherParams{
			ID:   randomFetcher.ID,
			Ip:   randomFetcher.Ip,
			Port: randomFetcher.Port,
		}, nil
	}
	return &FetcherParams{
		ID:   fetcher.ID,
		Ip:   fetcher.Ip,
		Port: fetcher.Port,
	}, nil
}

func GetLeastFullFetcher(ctx context.Context, db *orm.Queries) (*FetcherParams, error) {
	f, err := db.GetLeastFullFetcher(ctx)
	if err != nil {
		return nil, err
	}
	return &FetcherParams{
		ID:   f.ID,
		Ip:   f.Ip,
		Port: f.Port,
	}, nil
}

func GetMostFullFetcher(ctx context.Context, db *orm.Queries) (*FetcherParams, error) {
	f, err := db.GetMostFullFetcher(ctx)
	if err != nil {
		return nil, err
	}
	return &FetcherParams{
		ID:   f.ID,
		Ip:   f.Ip,
		Port: f.Port,
	}, nil
}
