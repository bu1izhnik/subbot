package bot

import (
	"context"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

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
	b.channelReposts.Mutex.Lock()
	b.channelReposts.List = make(map[tools.MessageConfig][]tools.RepostedTo)
	b.channelReposts.Mutex.Unlock()

	b.channelEdit.Mutex.Lock()
	b.channelEdit.List = make(map[tools.MessageConfig]string)
	b.channelEdit.Mutex.Unlock()
}

// IncreaseMsgCountForUser Returns true if user still not limited
func (b *Bot) IncreaseMsgCountForUser(userID int64) bool {
	b.usersLimits.Mutex.Lock()
	defer b.usersLimits.Mutex.Unlock()
	if _, ok := b.usersLimits.List[userID]; !ok {
		b.usersLimits.List[userID] = &tools.RateLimitConfig{
			MsgCnt:       1,
			LimitedUntil: time.Now().Add(-time.Duration(b.RateLimitCheckInterval) * time.Second),
		}
	} else {
		if time.Now().Before(b.usersLimits.List[userID].LimitedUntil) {
			return false
		}
		b.usersLimits.List[userID].MsgCnt++
		//log.Printf("New ratelimitcnt for %v: %v", userID, b.usersLimits.List[userID].MsgCnt)
		if b.usersLimits.List[userID].MsgCnt > b.RateLimitMaxMessages {
			b.usersLimits.List[userID].LimitedUntil = time.Now().Add(time.Duration(b.RateLimitTime) * time.Second)
			b.usersLimits.List[userID].MsgCnt = 0
			return false
		}
	}
	return true
}

func (b *Bot) checkForRateLimits() {
	b.usersLimits.Mutex.Lock()
	defer b.usersLimits.Mutex.Unlock()
	for userID, limit := range b.usersLimits.List {
		if time.Now().After(limit.LimitedUntil) && limit.MsgCnt == 0 {
			if limit.MsgCnt == 0 {
				delete(b.usersLimits.List, userID)
			} else {
				b.usersLimits.List[userID].MsgCnt = 0
				b.usersLimits.List[userID].LimitedUntil = time.Now().Add(-time.Duration(b.RateLimitCheckInterval) * time.Second)
			}
		} else {
			b.usersLimits.List[userID].MsgCnt = 0
		}
	}
}
