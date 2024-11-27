package fetcher

import (
	boltstor "github.com/gotd/contrib/bbolt"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lj "gopkg.in/natefinch/lumberjack.v2"
	"math/rand"
	"sync"

	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: cache messages for handling edits
// TODO: forward noforward messages by copying text

func Init(apiID int, apiHash string, botUsername string, mediaWait time.Duration) (*Fetcher, error) {
	f := &Fetcher{
		botID:       0,
		botHash:     0,
		botUsername: botUsername,
		sendChan:    make(chan *sendConfig, 1000),
		multiMediaQueue: AsyncMap[int64, *sendConfig]{
			List:  make(map[int64]*sendConfig),
			Mutex: sync.Mutex{},
		},
		mediaWaitTimer: mediaWait,
	}

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
	})

	session := telegram.FileSessionStorage{
		Path: "./session.json",
	}

	logWriter := zapcore.AddSync(&lj.Logger{
		Filename:   "./logs/log.json",
		MaxBackups: 3,
		MaxSize:    1, // megabytes
		MaxAge:     7, // days
	})
	logCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		logWriter,
		zap.DebugLevel,
	)
	lg := zap.New(logCore)
	defer func() { _ = lg.Sync() }()

	boltdb, err := bbolt.Open("./updates.bolt.db", 0666, nil)
	if err != nil {
		log.Fatalf("Error creating bolt db: %v", err)
	}
	updatesRecovery := updates.New(updates.Config{
		Handler: gaps, // using previous handler with peerDB
		Logger:  lg.Named("updates.recovery"),
		Storage: boltstor.NewStateStorage(boltdb),
	})

	client := telegram.NewClient(apiID, apiHash, telegram.Options{
		Logger:         lg,
		UpdateHandler:  updatesRecovery,
		SessionStorage: &session,
	})

	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		go f.handleNewMessage(ctx, update)
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

			gotForwardUpdate, err := f.client.API().MessagesForwardMessages(ctx, &tg.MessagesForwardMessagesRequest{
				FromPeer: &tg.InputPeerChannel{
					ChannelID:  send.forward.channelID,
					AccessHash: send.forward.accessHash,
				},
				ToPeer:   botPeer,
				ID:       send.forward.messageIDs,
				RandomID: getRandomIDs(len(send.forward.messageIDs)),
			})
			if err != nil {
				log.Printf(
					"Error forwarding message from %v: %v",
					send.forward.channelID,
					err,
				)
				continue
			}

			forwardUpdate, ok := gotForwardUpdate.(*tg.Updates)
			if !ok {
				log.Printf("Got incorrect type of update from forwarding: %T", gotForwardUpdate)
				continue
			}

			messageSentNotAsForward := false
			maxID := 0
			cnt := 0
			for _, update := range forwardUpdate.Updates {
				switch messageUpdate := update.(type) {
				case *tg.UpdateMessageID:
					maxID = max(maxID, messageUpdate.ID)
					cnt++
				case *tg.UpdateNewMessage:
					if messageSentNotAsForward {
						continue
					}
					message, ok := messageUpdate.Message.(*tg.Message)
					if ok {
						if _, ok := message.GetFwdFrom(); !ok {
							messageSentNotAsForward = true
						}
					}
				}
			}
			startID := maxID - cnt + 1

			if send.repost != nil {
				_, err = f.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer: botPeer,
					Message: fmt.Sprintf(
						"r %v %s %v",
						send.repost.toID,
						send.repost.toName,
						cnt,
					),
					ReplyTo: &tg.InputReplyToMessage{
						ReplyToMsgID: startID,
					},
					RandomID: rand.Int63(),
				})
				if err != nil {
					log.Printf("Error forwarding repost config: %v", err)
				}
			} else if send.edit != nil {
				// Temporarily off
				continue
			} else if messageSentNotAsForward { // Message sent no as forward so it needs additional info about channel in config
				_, err = f.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer: botPeer,
					Message: fmt.Sprintf(
						"w %v %s %v",
						send.forward.channelID,
						send.forward.channelName,
						cnt,
					),
					ReplyTo: &tg.InputReplyToMessage{
						ReplyToMsgID: startID,
					},
					RandomID: rand.Int63(),
				})
				if err != nil {
					log.Printf("Error forwarding weird post config: %v", err)
				}
			} else { // regular post
				_, err = f.client.API().MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
					Peer: botPeer,
					Message: fmt.Sprintf(
						"p %v",
						cnt,
					),
					ReplyTo: &tg.InputReplyToMessage{
						ReplyToMsgID: startID,
					},
					RandomID: rand.Int63(),
				})
				if err != nil {
					log.Printf("Error forwarding post config: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
		time.Sleep(interval)
	}
}
