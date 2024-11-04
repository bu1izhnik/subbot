package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/BulizhnikGames/subbot/fetcher/tools"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Fetcher struct {
	client  *telegram.Client
	gaps    *updates.Manager
	forward int64
}

func Init(apiID int, apiHash string, botChat int64) (*Fetcher, error) {
	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	client := telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: gaps})
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.AsNotEmpty()
		if !ok {
			return errors.New("unexpected message")
		}
		channelID, err := tools.GetChannelIDFromMessage(msg.String())
		if err != nil {
			return err
		}
		messageID, err := tools.GetMessageIDFromMessage(msg.String())
		if err != nil {
			return err
		}
		_, err = client.API().MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
			FromPeer: &tg.InputPeerChannelFromMessage{
				ChannelID: channelID,
				MsgID:     int(messageID),
			},
			ToPeer: &tg.InputPeerChat{
				ChatID: botChat,
			},
			ID: []int{int(messageID)},
		})
		return err
	})

	return &Fetcher{
		client:  client,
		gaps:    gaps,
		forward: botChat,
	}, nil
}

func (s *Fetcher) Run(phone string, password string, apiURL string, IP string, port string) error {
	codePrompt := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		log.Print("Enter code: ")
		code, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}

	flow := auth.NewFlow(
		auth.Constant(
			phone,
			password,
			auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{})

	return s.client.Run(context.Background(), func(ctx context.Context) error {
		if err := s.client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := s.client.Self(ctx)
		if err != nil {
			return err
		}

		jsonBody := []byte(fmt.Sprintf("{\"phone\": \"%s\" \"ip\": \"%s\" \"port\": \"%s\"}", phone, IP, port))
		bodyReader := bytes.NewReader(jsonBody)
		requestURL := apiURL + "/" + strconv.FormatInt(user.ID, 10)
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			return err
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			//return err
		}

		log.Printf("Scraper is %s:", user.Username)

		return s.gaps.Run(ctx, s.client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Gaps started")
			},
		})
	})
}

func (s *Fetcher) SubscribeToChannel(ctx context.Context, channelName string) (int64, int64, error) {
	res, err := s.client.API().ContactsResolveUsername(ctx, channelName)
	if err != nil {
		return 0, 0, err
	}
	channelID, err := tools.GetChannelIDFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	accessHash, err := tools.GetAccessHashFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	channel := tg.InputChannel{ChannelID: channelID, AccessHash: accessHash}
	_, err = s.client.API().ChannelsJoinChannel(ctx, &channel)
	return channelID, accessHash, err
}

/*func (s *Fetcher) GetUsername(ctx context.Context, channelID int64) error {
	channel := tg.InputChannel{ChannelID: channelID}
	res, err := s.client.API().ChannelsGetFullChannel(ctx, &channel)
	if err != nil {
		return err
	}
	log.Printf("Channel: %s", res.String())
	return nil
}*/
