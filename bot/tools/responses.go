package tools

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func SendErrorMessage(api *tgbotapi.BotAPI, message tgbotapi.MessageConfig) {
	_, err := api.Send(message)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func ResponseToCallback(api *tgbotapi.BotAPI, update tgbotapi.Update, newText string) error {
	if update.CallbackQuery == nil {
		return errors.New("no callback query to response with error")
	}

	groupID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID

	_, err := api.Send(tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      groupID,
			MessageID:   messageID,
			ReplyMarkup: nil,
		},
	})
	if err != nil {
		return err
	}

	_, err = api.Send(tgbotapi.NewEditMessageText(
		groupID,
		messageID,
		newText,
	))
	return err
}

func ResponseToCallbackLogError(api *tgbotapi.BotAPI, update tgbotapi.Update, newText string) {
	err := ResponseToCallback(api, update, newText)
	if err != nil {
		log.Printf("Error responding to callback query: %v", err)
	}
}
