package bot

import (
	"context"
	"errors"
	"github.com/BulizhnikGames/subbot/bot/tools"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
)

func (b *Bot) handleFromFetcher(ctx context.Context, update tgbotapi.Update) error {
	if update.Message == nil {
		return errors.New("message is not from fetcher: message empty")
	}

	if update.Message.ForwardFromChat == nil && update.Message.ForwardFrom == nil {
		return b.handleConfigMessage(ctx, update)
	}

	if update.Message.ForwardFromChat != nil {
		channelID, err := tools.GetChannelID(update.Message.ForwardFromChat.ID)
		if err != nil {
			return err
		}
		go b.tryUpdateChannelName(ctx, channelID, update.Message.ForwardFromChat.UserName)
		log.Printf("Got post: channel: ID: %v, Name: %s", channelID, update.Message.ForwardFromChat.UserName)
	}
	return nil
}

func (b *Bot) forwardPostToSubs(ctx context.Context, channelID int64, fetcherID int64, messageIDs *[]int, additional ...string) error {
	if len(*messageIDs) == 0 {
		return errors.New("no messages to forward")
	}

	groups, err := b.db.GetSubsOfChannel(ctx, channelID)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if len(additional) != 0 {
			_, err = b.api.Send(tgbotapi.NewMessage(group, additional[0]))
			if err != nil {
				log.Printf("Error sending additional mesasge to group %v: %v", group, err)
				continue
			}
		}

		if len(*messageIDs) == 1 {
			_, err := b.api.Send(tgbotapi.NewForward(group, fetcherID, (*messageIDs)[0]))
			if err != nil {
				log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
			}
		} else {
			err := b.forwardMessages(group, fetcherID, messageIDs)
			if err != nil {
				log.Printf("Error sending forward from channel %v to group %v: %v", channelID, group, err)
			}
		}
	}
	return nil
}

func (b *Bot) handleConfigMessage(ctx context.Context, update tgbotapi.Update) error {
	if len(update.Message.Text) > 2 && update.Message.ReplyToMessage != nil {
		replyMsg := update.Message.ReplyToMessage
		if update.Message.Text[0] == 'p' { // got post config (ex: "p messageCnt")
			cnt, err := strconv.Atoi(update.Message.Text[2:])
			if err != nil {
				return err
			}
			if replyMsg.ForwardFromChat == nil {
				return errors.New("error applying default post's config to message: it's not a forward")
			}
			channelID, err := tools.GetChannelID(replyMsg.ForwardFromChat.ID)
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				channelID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, cnt),
			)
		} else if update.Message.Text[0] == 'r' { // got repost message config (ex: "r channelID username messageCnt")
			rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				rep.To.ID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, rep.Cnt),
				"@"+rep.To.Name+" переслал сообщение:",
			)
		} else if update.Message.Text[0] == 'e' { // got edit message config (ex: "e cID1 mID2 username")
			// Temporarily off
			return nil
		} else if update.Message.Text[0] == 'w' { // got config for weird message (doesn't look like forwarded from chan, example - audio files) (ex: "w messageCnt")
			w, err := tools.GetValuesFromWeirdConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				w.Channel.ID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, w.Cnt),
				"Новое сообщение в канале @"+w.Channel.Name+":",
			)
		} else {
			return errors.New("message is not from fetcher: incorrect code in the beginning")
		}
	} else {
		return errors.New("message is not from fetcher: not forwarded and text length <= 2")
	}
}
