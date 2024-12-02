package fetcher

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

// This config sends to main bot when there is need to handle repost
type repostConfig struct {
	// ID of channel which reposted
	toID int64
	// Name of channel which reposted
	toName string
}

type forwardConfig struct {
	channelID   int64
	channelName string
	accessHash  int64
	messageIDs  []int
	idWithText  int
}

type sendConfig struct {
	forward   *forwardConfig
	repost    *repostConfig
	edit      bool
	noForward bool
}

type Fetcher struct {
	client          *telegram.Client
	redis           *redis.Client
	gaps            *updates.Manager
	sendChan        chan *sendConfig
	multiMediaQueue AsyncMap[int64, *sendConfig]
	mediaWaitTimer  time.Duration
	botUsername     string
	botPeer         *tg.InputPeerUser
}
