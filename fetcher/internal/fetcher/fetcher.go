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
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Fetcher struct {
	client *telegram.Client
	gaps   *updates.Manager
}

func Init(apiID int, apiHash string, botID int64, botHash int64) (*Fetcher, error) {
	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	client := telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: gaps})
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		msg, ok := update.Message.(*tg.Message)
		if !ok {
			log.Printf("Unexpected message type: %T", update.Message)
			return errors.New("unexpected message")
		}

		peer, ok := msg.PeerID.(*tg.PeerChannel)
		if !ok {
			log.Printf("Unexpected message's peer type: %T", msg.PeerID)
			return errors.New("unexpected peer")
		}

		getChannel, err := client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{
			&tg.InputChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: 0,
			},
		})
		if err != nil {
			log.Printf("Error getting channels access hash: %v", err)
			return err
		}
		channelData, ok := getChannel.(*tg.MessagesChats)
		if !ok {
			log.Printf("Unexpected channel type: %T", getChannel)
			return errors.New("unexpected channel")
		} else if channelData.Chats == nil {
			log.Printf("Error: empty channel")
			return errors.New("unexpected channel")
		}
		channel, ok := channelData.Chats[0].(*tg.Channel)
		if !ok {
			log.Printf("Unexpected channel chat type: %T", channelData.Chats[0])
			return errors.New("unexpected channel")
		}

		/*getBot, err := client.API().UsersGetFullUser(ctx, &tg.InputUser{
			UserID:     botChat,
			AccessHash: 0,
		})
		if err != nil {
			log.Println(err)
			return err
		}
		bot, ok := getBot.Chats[0].(*tg.Chat)
		if !ok {
			log.Println(err)
			return errors.New("unexpected bot")
		}

		log.Printf("%v, %v \n %v, %v", peer.ChannelID, channel.AccessHash, bot.ID)*/

		/*resolved, err := client.API().ContactsResolveUsername(ctx, "stonesubbot")
		if err != nil {
			log.Printf("failed to resolve username: %v", err)
			return err
		}

		var botPeer tg.InputPeerClass
		if len(resolved.Users) > 0 {
			user := resolved.Users[0]
			if u, ok := user.(*tg.User); ok {
				botPeer = &tg.InputPeerUser{
					UserID:     u.ID,
					AccessHash: u.AccessHash,
				}
			}
		} else {
			return fmt.Errorf("could not resolve bot username")
		}

		log.Printf("%v, %v", botPeer.(*tg.InputPeerUser).UserID, botPeer.(*tg.InputPeerUser).AccessHash)*/

		messageID, err := tools.GetMessageIDFromMessage(msg.String())
		if err != nil {
			return err
		}

		_, err = client.API().MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
			FromPeer: &tg.InputPeerChannel{
				ChannelID:  peer.ChannelID,
				AccessHash: channel.AccessHash,
			},
			ToPeer: &tg.InputPeerUser{
				UserID:     botID,
				AccessHash: botHash,
			},
			ID:       []int{messageID},
			RandomID: []int64{rand.Int63()},
		})
		if err != nil {
			log.Printf("Error forwarding messages: %v", err)
		}
		return err
	})

	return &Fetcher{
		client: client,
		gaps:   gaps,
	}, nil
}

func (f *Fetcher) Run(phone string, password string, apiURL string, IP string, port string) error {
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

	return f.client.Run(context.Background(), func(ctx context.Context) error {
		if err := f.client.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := f.client.Self(ctx)
		if err != nil {
			return err
		}

		jsonBody := []byte(fmt.Sprintf("{\"phone\": \"%s\", \"ip\": \"%s\", \"port\": \"%s\"}", phone, IP, port))
		bodyReader := bytes.NewReader(jsonBody)
		requestURL := apiURL + "/" + strconv.FormatInt(user.ID, 10)
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			return err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[POSSIBLE ERROR]: couldn't register fetcher by API request (register manualy or restart): %v", err)
		} else if res.StatusCode != http.StatusOK {
			log.Printf("[POSSIBLE ERROR]: couldn't register fetcher by API request (register manualy or restart). HTTP status code: %v", res.StatusCode)
		}

		log.Printf("Scraper is %s:", user.Username)

		return f.gaps.Run(ctx, f.client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Println("Gaps started")
			},
		})
	})
}

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
	channelID, err := tools.GetChannelIDFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	accessHash, err := tools.GetAccessHashFromChannel(res.Chats[0].String())
	if err != nil {
		return 0, 0, err
	}
	return channelID, accessHash, err
}
