package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	"log"
	"strconv"
	"strings"
)

func (f *Fetcher) handleNewMessage(ctx context.Context, update *tg.UpdateNewChannelMessage) error {
	channel, msg, err := f.getChannelAndMessageInfo(ctx, update.Message)
	if err != nil {
		log.Printf("Error handling new message in channel: %v", err)
		return err
	}

	var repostCfg *repostConfig
	forwardCfg := &forwardConfig{
		channelID:   channel.ID,
		channelName: channel.Username,
		accessHash:  channel.AccessHash,
		messageIDs:  []int{msg.ID},
	}

	if _, ok := msg.GetFwdFrom(); ok {
		repostCfg = &repostConfig{
			toID:   channel.ID,
			toName: channel.Username,
		}
	}

	if msg.Message != "" {
		forwardCfg.idWithText = msg.ID
	}

	sendCfg := sendConfig{
		repost:  repostCfg,
		forward: forwardCfg,
	}

	if msg.GroupedID != 0 {
		f.handleMultimedia(
			msg.GroupedID,
			&sendCfg,
		)
		return nil
	}

	f.sendChan <- &sendCfg

	return nil
}

func (f *Fetcher) handleMultimedia(groupID int64, sendCfg *sendConfig) {
	f.multiMediaQueue.Mutex.Lock()
	if f.multiMediaQueue.List[groupID] != nil {
		f.multiMediaQueue.List[groupID].forward.messageIDs =
			append(f.multiMediaQueue.List[groupID].forward.messageIDs, sendCfg.forward.messageIDs...)
		if sendCfg.forward.idWithText != 0 {
			f.multiMediaQueue.List[groupID].forward.idWithText = sendCfg.forward.idWithText
		}
		f.multiMediaQueue.Mutex.Unlock()
	} else {
		f.multiMediaQueue.List[groupID] = sendCfg
		f.multiMediaQueue.Mutex.Unlock()
		f.waitForOtherMediaInGroup(groupID)
	}
}

func (f *Fetcher) handleEdit(ctx context.Context, update *tg.UpdateEditChannelMessage) error {
	channel, msg, err := f.getChannelAndMessageInfo(ctx, update.Message)
	if err != nil {
		return err
	}
	// Avoid posts with giveaways (each press on button in these posts is new edit update)
	if msg.ReplyMarkup != nil {
		return nil
	}

	channelIDStr := strconv.FormatInt(channel.ID, 10)
	messageIDStr := strconv.Itoa(msg.ID)
	//log.Printf("message:" + channelIDStr + ":" + messageIDStr)
	oldMsgIDStr, err := f.redis.Get(ctx, "message:"+channelIDStr+":"+messageIDStr).Result()
	if err != nil {
		// handle cases where there is no data in DB
		if strings.Contains(err.Error(), "nil") {
			return nil
		}
		return err
	}
	oldMsgID, err := strconv.Atoi(oldMsgIDStr)
	if err != nil {
		return err
	}

	getMessageInBotChat, err := f.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     f.botPeer,
		OffsetID: oldMsgID + 1,
	})
	messageInBotChat, ok := getMessageInBotChat.(*tg.MessagesMessagesSlice)
	if !ok {
		return errors.New(fmt.Sprintf("got unexpected result type from bot's chat: %T", getMessageInBotChat))
	}
	if len(messageInBotChat.Messages) == 0 {
		return errors.New("got empty list of messages from bot's chat")
	}

	messageCopy, ok := messageInBotChat.Messages[0].(*tg.Message)
	if !ok {
		return errors.New(fmt.Sprintf("got unexpected message type from bot's chat: %T", messageInBotChat.Messages[0]))
	}

	if messageCopy.Message == msg.Message {
		return nil
	}

	IDs := make([]int, 1, 10)
	IDs[0] = msg.ID
	if msg.GroupedID != 0 {
		getOldMessages, err := f.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			},
			OffsetID: msg.ID + 10,
			Limit:    19,
		})
		if err != nil {
			return err
		}

		oldMessages, ok := getOldMessages.(*tg.MessagesChannelMessages)
		if !ok {
			return errors.New(fmt.Sprintf("got unexpected result type from channel: %T", oldMessages))
		}
		if len(oldMessages.Messages) == 0 {
			return errors.New("got empty list of messages from channel")
		}

		for _, message := range oldMessages.Messages {
			oldMessage, ok := message.(*tg.Message)
			log.Printf("id: %v", oldMessage.ID)
			if ok && oldMessage.GroupedID == msg.GroupedID && oldMessage.ID != msg.ID {
				IDs = append(IDs, oldMessage.ID)
			}
		}
	}

	//log.Printf("id to send: %v", IDs[0])

	f.sendChan <- &sendConfig{
		forward: &forwardConfig{
			channelID:   channel.ID,
			accessHash:  channel.AccessHash,
			channelName: channel.Username,
			messageIDs:  IDs,
			idWithText:  msg.ID,
		},
		edit: true,
	}
	return nil
}
