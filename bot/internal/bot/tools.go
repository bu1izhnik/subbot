package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/db/orm"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

func (b *Bot) forwardMessages(toChat int64, fetcherChat int64, messageIDs *[]int) error {
	toChatStr := strconv.FormatInt(toChat, 10)
	fromChatStr := strconv.FormatInt(fetcherChat, 10)
	messageIDsStr := strings.Builder{}
	for i, id := range *messageIDs {
		messageIDsStr.WriteString(strconv.Itoa(id))
		if i != len(*messageIDs)-1 {
			messageIDsStr.WriteString(", ")
		}
	}
	//messageIDsStr.WriteString("367")
	//b.api.Send(tgbotapi.NewForward(toChat, fetcherChat, 367))
	url := fmt.Sprintf("https://api.telegram.org/bot%s/forwardMessages", b.api.Token)
	jsonBody := []byte(
		fmt.Sprintf(
			"{\"chat_id\": %s, \"from_chat_id\": %s, \"message_ids\": [ %s ]}",
			toChatStr,
			fromChatStr,
			messageIDsStr.String(),
		),
	)
	bodyReader := bytes.NewReader(jsonBody)

	//log.Printf("req body: %s", jsonBody)

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var apiResp tgbotapi.APIResponse
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &apiResp)
	if err != nil {
		return err
	}

	//log.Printf("Reps: %s", string(data))

	return nil
}
