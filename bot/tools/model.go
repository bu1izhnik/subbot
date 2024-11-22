package tools

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
	"time"
)

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

type FetcherParams struct {
	ID   int64
	Ip   string
	Port string
}

type MessageConfig struct {
	ChannelID int64
	MessageID int
}

type RepostedTo struct {
	ChannelID   int64
	ChannelName string
}

type RateLimitConfig struct {
	MsgCnt       int64
	LimitedUntil time.Time
}

type MultiMediaConfig struct {
	FromFetcherChat int64
	FromChannel     int64
	IDs             [10]int
	Cnt             int
	GotMaxMedia     chan struct{}
}

type GetFetcherRequest func(ctx context.Context, db *orm.Queries) (*FetcherParams, error)

type Command func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error
