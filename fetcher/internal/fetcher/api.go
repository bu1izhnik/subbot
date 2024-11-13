package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
)

func (f *Fetcher) SubscribeToChannel(ctx context.Context, channelName string) (int64, int64, error) {
	channelID, accessHash, canForward, err := f.GetChannelInfo(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	// maybe let user sub and forward messages like reposts or edits
	if !canForward {
		return 0, 0, errors.New("can't subscribe to channel which messages can't be forwarded")
	}
	channel := tg.InputChannel{ChannelID: channelID, AccessHash: accessHash}
	_, err = f.client.API().ChannelsJoinChannel(ctx, &channel)
	return channelID, accessHash, err
}

// GetChannelInfo returns channel's id, access hash, can you forward from it and error (in this order)
func (f *Fetcher) GetChannelInfo(ctx context.Context, channelName string) (int64, int64, bool, error) {
	res, err := f.client.API().ContactsResolveUsername(ctx, channelName)
	if err != nil {
		return 0, 0, false, err
	}
	if len(res.Chats) == 0 {
		return 0, 0, false, errors.New("not a channel: got 0 chats by resolving")
	}
	if channel, ok := res.Chats[0].(*tg.Channel); ok {
		if channel.Gigagroup || channel.Megagroup {
			return 0, 0, false, errors.New("not a channel: invalid chat type - super/mega group")
		}
		return channel.ID, channel.AccessHash, !channel.Noforwards, nil
	} else {
		return 0, 0, false, errors.New(fmt.Sprintf("not a channel: invalid chat type (%T)", res.Chats[0]))
	}
}
