package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) handleFromFetcher(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("message is not from fetcher: message empty")
	}

	/*log.Printf("%v", update.Message.MessageID)
	if update.Message.ReplyToMessage != nil {
		log.Printf("reply to id: %v", update.Message.ReplyToMessage.MessageID)
	}*/

	if update.Message.ForwardFromChat == nil && update.Message.ForwardFrom == nil {
		return b.handleConfigMessage(ctx, update)
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

	/*forwardMessageID := update.Message.ForwardFromMessageID
	msgCfg := tools.MessageConfig{
		ChannelID: chatID,
		MessageID: forwardMessageID,
	}*/

	/*ok, err := b.tryHandleEdit(ctx, update, msgCfg, chatID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	ok, err = b.tryHandleRepost(ctx, update, msgCfg, chatID)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}*/

	if update.Message.MediaGroupID != "" {
		mediaGroup, _ := strconv.ParseInt(
			update.Message.MediaGroupID,
			10,
			64,
		)
		b.queueMultiMedia(
			update.Message.MessageID,
			chatID,
			update.Message.Chat.ID,
			mediaGroup,
		)
		return nil
	}

	go b.tryUpdateChannelName(ctx, chatID, update.Message.ForwardFromChat.UserName)

	return b.forwardPostToSubs(
		ctx,
		chatID,
		update.Message.Chat.ID,
		&[]int{update.Message.MessageID},
	)
}

func (b *Bot) forwardPostToSubs(ctx context.Context, channelID int64, fetcherID int64, messageIDs *[]int, additional ...string) error {
	if len(*messageIDs) == 0 {
		return errors.New("no messages to forward")
	}

	groups, err := b.db.GetSubsOfChannel(ctx, channelID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if len(additional) != 0 {
			_, err = b.api.Send(tgbotapi.NewMessage(group, additional[0]))
			if err != nil {
				log.Printf("Error sending additional mesasge to group %v: %v", group, err)
				continue
			}
		}

		if len(*messageIDs) == 1 {
			_, err := b.api.Send(tgbotapi.NewForward(group, fetcherID, (*messageIDs)[0]))
			if err != nil {
				log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
			}
		} else {
			err := b.forwardMessages(group, fetcherID, messageIDs)
			if err != nil {
				log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
			}
		}
	}
	return nil
}

func (b *Bot) handleConfigMessage(ctx context.Context, update tgbotapi.Update) error {
	if len(update.Message.Text) > 2 || update.Message.ReplyToMessage != nil {
		if update.Message.Text[0] == 'r' { // got repost message config (ex: "r channelID username messageCnt")
			/*cfg, rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			b.channelReposts.Mutex.Lock()
			if b.channelReposts.List[*cfg] == nil {
				b.channelReposts.List[*cfg] = make([]tools.Repost, 0)
			}
			b.channelReposts.List[*cfg] = append(b.channelReposts.List[*cfg], *rep)
			//log.Printf("added to reposts from: %+v to: %+v (%v)", *cfg, *rep, len(b.channelReposts.reposts[*cfg]))
			b.channelReposts.Mutex.Unlock()
			return nil*/
			rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			groupID, err := strconv.ParseInt(update.Message.ReplyToMessage.MediaGroupID, 10, 64)
			if err != nil {
				return err
			}
			b.multiMediaQueue.Mutex.Lock()
			b.multiMediaQueue.List[groupID].WasRepost <- struct{}{}
			b.multiMediaQueue.Mutex.Unlock()
			IDs := make([]int, rep.Cnt)
			for i := 0; i < rep.Cnt; i++ {
				IDs[i] = update.Message.ReplyToMessage.MessageID + i
			}
			return b.forwardPostToSubs(
				ctx,
				rep.ChannelID,
				update.Message.Chat.ID,
				&IDs,
				"@"+rep.ChannelName+" переслал сообщение:",
			)
		} else if update.Message.Text[0] == 'e' { // got edit message config (ex: "e cID1 mID2 username")
			cfg, channelName, err := tools.GetValuesFromEditConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			b.channelEdit.Mutex.Lock()
			b.channelEdit.List[*cfg] = channelName
			b.channelEdit.Mutex.Unlock()
			//log.Printf("setted edit: %+v = %s", *cfg, channelName)
			return nil
		} else {
			return errors.New("message is not from fetcher: incorrect code in the beginning")
		}
	} else {
		return errors.New("message is not from fetcher: not forwarded and text length <= 2")
	}
}

func (b *Bot) tryHandleEdit(ctx context.Context, update tgbotapi.Update, msgCfg tools.MessageConfig, channelID int64) (bool, error) {
	b.channelEdit.Mutex.Lock()
	if channelName, ok := b.channelEdit.List[msgCfg]; ok {
		delete(b.channelEdit.List, msgCfg)
		b.channelEdit.Mutex.Unlock()

		err := b.forwardPostToSubs(
			ctx,
			channelID,
			update.Message.Chat.ID,
			&[]int{update.Message.MessageID},
			"@"+channelName+" отредактировал сообщение:",
		)
		return true, err
	} else {
		b.channelEdit.Mutex.Unlock()
		return false, nil
	}
}

func (b *Bot) tryHandleRepost(ctx context.Context, update tgbotapi.Update, msgCfg tools.MessageConfig, chatID int64) (bool, error) {
	b.channelReposts.Mutex.Lock()
	if targets, ok := b.channelReposts.List[msgCfg]; ok {
		delete(b.channelReposts.List, msgCfg)
		b.channelReposts.Mutex.Unlock()

		for _, target := range targets {
			err := b.forwardPostToSubs(
				ctx,
				target.ChannelID,
				update.Message.Chat.ID,
				&[]int{update.Message.MessageID},
				"@"+target.ChannelName+" переслал сообщение:",
			)

			if err != nil {
				if len(targets) > 0 {
					log.Printf(
						"Could not repost message from channel %v, to channel (%v, %s): %v",
						chatID,
						target.ChannelID,
						target.ChannelName,
						err,
					)
					continue
				} else {
					return true, errors.New(
						fmt.Sprintf(
							"Could not repost message from channel %v, to channel (%v, %s): %v",
							chatID,
							target.ChannelID,
							target.ChannelName,
							err,
						),
					)
				}
			}
		}
		return true, nil
	} else {
		b.channelReposts.Mutex.Unlock()
		return false, nil
	}
}
