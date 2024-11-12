package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
)

func (f *Fetcher) SubscribeToChannel(ctx context.Context, channelName string) (int64, int64, error) {
	channelID, accessHash, err := f.GetChannelInfo(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	channel := tg.InputChannel{ChannelID: channelID, AccessHash: accessHash}
	_, err = f.client.API().ChannelsJoinChannel(ctx, &channel)
	return channelID, accessHash, err
}

func (f *Fetcher) GetChannelInfo(ctx context.Context, channelName string) (int64, int64, error) {
	res, err := f.client.API().ContactsResolveUsername(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	if len(res.Chats) == 0 {
		return 0, 0, errors.New("not a channel: got 0 chats by resolving")
	}
	if channel, ok := res.Chats[0].(*tg.Channel); ok {
		return channel.ID, channel.AccessHash, nil
	} else {
		return 0, 0, errors.New(fmt.Sprintf("not a channel: invalid chat type (%T)", res.Chats[0]))
	}
}
