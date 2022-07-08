package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b := bot.New("...", opts...)
	fmt.Println("bot started")

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		fmt.Printf("callback id %d\n%s <- %s \n%#v\n", update.ID, update.CallbackQuery.Message.Text, update.CallbackQuery.Data, update.CallbackQuery.Message.ReplyMarkup)

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}

		if update.CallbackQuery.Message.ReplyMarkup != nil {
			switch c := update.CallbackQuery.Message.ReplyMarkup.(type) {
			case map[string]interface{}:
				// fmt.Printf("Item %#v\n", c["inline_keyboard"])
				buttons := []models.InlineKeyboardButton{}

				for _, v := range c["inline_keyboard"].([]interface{}) {
					subitems := v.([]interface{})
					for _, i := range subitems {
						fmt.Printf("SubItem %s => %s\n", i.(map[string]interface{})["callback_data"], i.(map[string]interface{})["text"])
						text := i.(map[string]interface{})["text"].(string)
						callbackData := i.(map[string]interface{})["callback_data"].(string)

						if callbackData == update.CallbackQuery.Data {
							if strings.HasPrefix(callbackData, "busy-") {
								callbackData = strings.Replace(callbackData, "busy-", "free-", 1)
								text = strings.Replace(text, "🟢", "🛑", 1)
							} else {
								callbackData = strings.Replace(callbackData, "free-", "busy-", 1)
								text = strings.Replace(text, "🛑", "🟢", 1)
							}
						}

						buttons = append(buttons, models.InlineKeyboardButton{Text: text, CallbackData: callbackData})
					}
				}

				kb.InlineKeyboard = [][]models.InlineKeyboardButton{buttons}
			default:
				// fmt.Printf("N%T\n", c)
			}
		}

		// kb = &models.InlineKeyboardMarkup{
		// 	InlineKeyboard: [][]models.InlineKeyboardButton{
		// 		{
		// 			{Text: "🟢 testing-id", CallbackData: "busy-testing-id"},
		// 			{Text: "🛑 testing-1", CallbackData: "free-testing-1"},
		// 			{Text: "🟢 testing-2", CallbackData: "busy-testing-2"},
		// 		},
		// 	},
		// }

		editedMessage := &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.ID,
			Text:        update.CallbackQuery.Message.Text,
			ReplyMarkup: kb,
		}

		b.EditMessageText(ctx, editedMessage)

		return
	}

	fmt.Printf("message %#v\n", update)

	// echo
	// b.SendMessage(ctx, &bot.SendMessageParams{
	// 	ChatID: update.Message.Chat.ID,
	// 	Text:   update.Message.Text,
	// })
}
