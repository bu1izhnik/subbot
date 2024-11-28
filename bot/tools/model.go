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

type ChannelInfo struct {
	ID   int64
	Name string
}

type RepostConfig struct {
	To  ChannelInfo
	Cnt int
}

type EditConfig struct {
	ChannelName string
	Cnt         int
}

type WeirdConfig struct {
	Channel ChannelInfo
	Cnt     int
}

type RateLimitConfig struct {
	MsgCnt       int64
	LimitedUntil time.Time
}

type GetFetcherRequest func(ctx context.Context, db *orm.Queries) (*FetcherParams, error)

type Command func(ctx context.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) error
