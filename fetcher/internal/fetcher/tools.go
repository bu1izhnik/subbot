package fetcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
)

func (f *Fetcher) getChannelAndMessageInfo(ctx context.Context, e tg.Entities, message tg.MessageClass) (*tg.Channel, *tg.Message, error) {
	msg, ok := message.(*tg.Message)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected message type %T:", message))
	}

	peer, ok := msg.PeerID.(*tg.PeerChannel)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("unexpected peer type: %T", msg.PeerID))
	}

	return e.Channels[peer.ChannelID], msg, nil
}

func (f *Fetcher) setBotHashAndID(ctx context.Context) error {
	resolved, err := f.client.API().ContactsResolveUsername(ctx, f.botUsername)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to resolve username of bot (%v) to foward message: %v", f.botUsername, err))
	}

	if len(resolved.Users) > 0 {
		user := resolved.Users[0]
		if u, ok := user.(*tg.User); ok {
			f.botID = u.ID
			f.botHash = u.AccessHash
		} else {
			return errors.New(fmt.Sprintf("failed to resolve username of bot (%v): not a user", f.botUsername))
		}
	} else {
		return errors.New(fmt.Sprintf("failed to resolve username of bot (%v): resolving returned 0 users", f.botUsername))
	}

	return nil
}
