package fetcher

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"sync"
	"time"
)

type AsyncMap[K comparable, V any] struct {
	Mutex sync.Mutex
	List  map[K]V
}

// This config sends to main bot when there is need to handle edit
type editConfig struct {
	// ID of channel in which edit was done
	channelID int64
	// ID of message which was edited
	messageID int
	// Name of channel in which edit was done
	channelName string
}

// This config sends to main bot when there is need to handle repost
type repostConfig struct {
	// ID of channel from which repost was done
	//fromID int64
	// ID of message in channel from which repost was done
	//messageIDs []int
	// ID of channel which reposted
	toID int64
	// Name of channel which reposted
	toName string
}

type forwardConfig struct {
	channelID  int64
	accessHash int64
	messageIDs []int
}

type sendConfig struct {
	edit    *editConfig
	repost  *repostConfig
	forward *forwardConfig
}

type Fetcher struct {
	client          *telegram.Client
	gaps            *updates.Manager
	sendChan        chan *sendConfig
	multiMediaQueue AsyncMap[int64, *sendConfig]
	mediaWaitTimer  time.Duration
	botUsername     string
	botID           int64
	botHash         int64
}
