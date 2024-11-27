package fetcher

import (
	"context"
	"github.com/gotd/td/tg"
	"log"
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

	sendCfg := sendConfig{
		repost:  repostCfg,
		forward: forwardCfg,
		edit:    nil,
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
		f.multiMediaQueue.Mutex.Unlock()
	} else {
		f.multiMediaQueue.List[groupID] = sendCfg
		f.multiMediaQueue.Mutex.Unlock()
		f.waitForOtherMediaInGroup(groupID)
	}
}
