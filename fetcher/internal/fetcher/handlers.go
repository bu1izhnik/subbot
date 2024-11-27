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
		channelID:  channel.ID,
		accessHash: channel.AccessHash,
		messageIDs: []int{msg.ID},
	}

	/*if fwd, ok := msg.GetFwdFrom(); ok {
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
			//fromID: originalChatID,
			//messageIDs: []int{originalMessageID},
			toID:   channel.ID,
			toName: channel.Username,
		}
	}*/

	if _, ok := msg.GetFwdFrom(); ok {
		repostCfg = &repostConfig{
			//fromID: originalChatID,
			//messageIDs: []int{originalMessageID},
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
		/*if f.multiMediaQueue.List[groupID].repost != nil {
			f.multiMediaQueue.List[groupID].repost.messageIDs =
				append(f.multiMediaQueue.List[groupID].repost.messageIDs, sendCfg.repost.messageIDs...)
		}*/
		log.Printf("Added new photo to group: %v", groupID)
		f.multiMediaQueue.Mutex.Unlock()
	} else {
		f.multiMediaQueue.List[groupID] = sendCfg
		f.multiMediaQueue.Mutex.Unlock()
		f.waitForOtherMediaInGroup(groupID)
	}
}
