package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
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
	"time"
)

// TODO: cache messages for handling edits
// TODO: forward noforward messages by copying text

// This config sends to main bot when there is need to handle edit
type editConfig struct {
	// ID of channel in which edit was done
	channelID int64
	// ID of message which was edited
	messageID int
	// Name of channel in which edit was done
	channelName string
}

// This config sends to main bot when there is need to handle repost
type repostConfig struct {
	// ID of channel from which repost was done
	fromID int64
	// ID of message in channel from which repost was done
	messageID int
	// ID of channel which reposted
	toID int64
	// Name of channel which reposted
	toName string
}

type forwardConfig struct {
	channelID  int64
	accessHash int64
	messageID  int
}

type sendConfig struct {
	edit    *editConfig
	repost  *repostConfig
	forward *forwardConfig
}

type Fetcher struct {
	client      *telegram.Client
	gaps        *updates.Manager
	sendChan    chan *sendConfig
	botUsername string
	botID       int64
	botHash     int64
}

func Init(apiID int, apiHash string, botUsername string) (*Fetcher, error) {
	f := &Fetcher{
		botID:       0,
		botHash:     0,
		botUsername: botUsername,
		sendChan:    make(chan *sendConfig, 1000),
	}

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})
	session := telegram.FileSessionStorage{
		Path: "./session.json",
	}
	client := telegram.NewClient(apiID, apiHash, telegram.Options{UpdateHandler: gaps, SessionStorage: &session})

	// edit config temporarily turned off because of wierd behaviour
	/*d.OnEditChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateEditChannelMessage) error {
		channel, msg, err := f.getChannelAndMessageInfo(ctx, update.Message)
		if err != nil {
			log.Printf("Error handling edited message in channel: %v", err)
			return err
		}

		// If message has reply markup (ex: giveaway) it will be seen as edited each time someone presses button, same with reactions
		if msg.ReplyMarkup != nil || len(msg.Reactions.Results) != 0 {
			return nil
		}

		f.sendChan <- &sendConfig{
			edit: &editConfig{
				channelID:   channel.ID,
				messageID:   msg.ID,
				channelName: channel.Username,
			},
			forward: &forwardConfig{
				channelID:  channel.ID,
				accessHash: channel.AccessHash,
				messageID:  msg.ID,
			},
			repost: nil,
		}
		return nil
	})*/

	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		channel, msg, err := f.getChannelAndMessageInfo(ctx, update.Message)
		if err != nil {
			log.Printf("Error handling new message in channel: %v", err)
			return err
		}

		var repostCfg *repostConfig
		forwardCfg := &forwardConfig{
			channelID:  channel.ID,
			accessHash: channel.AccessHash,
			messageID:  msg.ID,
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

		f.sendChan <- &sendConfig{
			repost:  repostCfg,
			forward: forwardCfg,
			edit:    nil,
		}

		return nil
	})

	f.client = client
	f.gaps = gaps
	return f, nil
}

func (f *Fetcher) Run(phone string, password string, apiURL string, IP string, port string) error {
	go f.tick(context.Background(), 500*time.Millisecond)

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

func (f *Fetcher) tick(ctx context.Context, interval time.Duration) {
	for {
		select {
		case send := <-f.sendChan:
			if send.forward == nil {
				log.Printf("Send config must contain forward config")
				continue
			}

			if f.botID == 0 || f.botHash == 0 {
				if err := f.setBotHashAndID(ctx); err != nil {
					log.Printf("Error setting bot ID and hash: %v", err)
					continue
				}
			}

			botPeer := &tg.InputPeerUser{
				UserID:     f.botID,
				AccessHash: f.botHash,
			}

			if send.edit != nil {
				_, err := f.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer:     botPeer,
					Message:  fmt.Sprintf("e %v %v %s", send.edit.channelID, send.edit.messageID, send.edit.channelName),
					RandomID: int64(rand.Int31()),
				})
				if err != nil {
					log.Printf("Error sending support message for edit: %v", err)
					continue
				}
			} else if send.repost != nil {
				_, err := f.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer:     botPeer,
					Message:  fmt.Sprintf("r %v %v %v %s", send.repost.fromID, send.repost.messageID, send.repost.toID, send.repost.toName),
					RandomID: int64(rand.Int31()),
				})
				if err != nil {
					log.Printf("Error sending support message for repost: %v", err)
					continue
				}
			}

			_, err := f.client.API().MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
				FromPeer: &tg.InputPeerChannel{
					ChannelID:  send.forward.channelID,
					AccessHash: send.forward.accessHash,
				},
				ToPeer:   botPeer,
				ID:       []int{send.forward.messageID},
				RandomID: []int64{rand.Int63()},
			})
			if err != nil {
				log.Printf(
					"Error forwarding message from %v: %v",
					send.forward.channelID,
					err,
				)
			}
		case <-ctx.Done():
			return
		}
		time.Sleep(interval)
	}
}
