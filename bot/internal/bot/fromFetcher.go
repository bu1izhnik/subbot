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
		} else if update.Message.Text[0] == 'r' { // got repost message config (ex: "r channelID channelName messageCnt")
			rep, err := tools.GetValuesFromRepostConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				rep.To.ID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, rep.Cnt),
				"@"+rep.To.Name+" переслал сообщение",
			)
		} else if update.Message.Text[0] == 'e' { // got edit message config (ex: "e channelName messageCnt")
			edit, err := tools.GetValuesFromEditConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			channelID, err := tools.GetChannelID(replyMsg.ForwardFromChat.ID)
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				channelID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, edit.Cnt),
				"@"+edit.ChannelName+" отредактировал сообщение",
			)
		} else if update.Message.Text[0] == 'n' { // got config for not forwardeble message (example - audio files, copied messages from channels with banned forwards) (ex: "n channelName messageCnt")
			w, err := tools.GetValuesFromNotForwardConfig(update.Message.Text[2:])
			if err != nil {
				return err
			}
			return b.forwardPostToSubs(
				ctx,
				w.Channel.ID,
				update.Message.Chat.ID,
				tools.GetIDs(replyMsg.MessageID, w.Cnt),
				"Новое сообщение в канале @"+w.Channel.Name,
			)
		} else {
			return errors.New("message is not from fetcher: incorrect code in the beginning")
		}
	} else {
		if update.Message.ReplyToMessage != nil {
			return errors.New("incorrect config")
		} else {
			return nil
		}
	}
}
