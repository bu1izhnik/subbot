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

type FetcherParams struct {
	ID   int64
	Ip   string
	Port string
}

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

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
	if username[0] == '@' {
		return username[1:]
	} else if strings.HasPrefix(username, "t.me/") {
		return username[5:]
	} else if strings.HasPrefix(username, "https://t.me/") {
		return username[13:]
	}
	return username
}

func SendWithErrorLogging(api *tgbotapi.BotAPI, message tgbotapi.MessageConfig) {
	message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	_, err := api.Send(message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
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
