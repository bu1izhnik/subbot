package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	"log"
)

func (f *Fetcher) handleNewMessage(ctx context.Context, update *tg.UpdateNewChannelMessage) error {
	channel, msg, err := f.getChannelAndMessageInfo(ctx, update.Message)
	if err != nil {
		log.Printf("Error handling new message in channel: %v", err)
		return err
	}

	if msg.GroupedID != 0 {
		f.multiMediaQueue.Mutex.Lock()
		if f.multiMediaQueue.List[msg.GroupedID] != nil {
			f.multiMediaQueue.List[msg.GroupedID].forward.messageIDs =
				append(f.multiMediaQueue.List[msg.GroupedID].forward.messageIDs, msg.ID)
			log.Printf("Added new photo to group: %v", msg.GroupedID)
			f.multiMediaQueue.Mutex.Unlock()
			return nil
		}
		f.multiMediaQueue.Mutex.Unlock()
	}

	var repostCfg *repostConfig
	forwardCfg := &forwardConfig{
		channelID:  channel.ID,
		accessHash: channel.AccessHash,
		messageIDs: []int{msg.ID},
	}

	if fwd, ok := msg.GetFwdFrom(); ok {
		var originalChatID int64
		originalMessageID := fwd.ChannelPost
		// No chat peer support now
		switch p := fwd.FromID.(type) {
		case *tg.PeerChannel:
			originalChatID = p.ChannelID
		case *tg.PeerUser:
			originalChatID = p.UserID
		default:
			log.Printf("Can't handle repost: unexpected type of original peer: %T", fwd.FromID)
			return errors.New(fmt.Sprintf("can't handle repost: unexpected type of original peer: %T", fwd.FromID))
		}

		repostCfg = &repostConfig{
			fromID:    originalChatID,
			messageID: originalMessageID,
			toID:      channel.ID,
			toName:    channel.Username,
		}
	}

	if msg.GroupedID != 0 {
		f.handleMultimedia(
			msg.GroupedID,
			&sendConfig{
				repost:  repostCfg,
				forward: forwardCfg,
				edit:    nil,
			},
		)
		return nil
	}

	f.sendChan <- &sendConfig{
		repost:  repostCfg,
		forward: forwardCfg,
		edit:    nil,
	}

	return nil
}

func (f *Fetcher) handleMultimedia(groupID int64, sendCfg *sendConfig) {
	f.multiMediaQueue.Mutex.Lock()
	if f.multiMediaQueue.List[groupID] != nil {
		log.Printf("Added new photo to group: %v", groupID)
		f.multiMediaQueue.List[groupID].forward.messageIDs =
			append(f.multiMediaQueue.List[groupID].forward.messageIDs, sendCfg.forward.messageIDs...)
		f.multiMediaQueue.Mutex.Unlock()
	} else {
		f.multiMediaQueue.List[groupID] = sendCfg
		f.multiMediaQueue.Mutex.Unlock()
		f.waitForOtherMediaInGroup(groupID)
	}
}
