package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const ConfigFileName = "/data/options.json"

// Config ...
type Config struct {
	Token string `json:"TOKEN"`
}

func main() {
	token := ""
	var initFromFile = false

	if _, err := os.Stat(ConfigFileName); err == nil {
		jsonFile, err := os.Open(ConfigFileName)
		if err == nil {
			config := &Config{}

			byteValue, _ := io.ReadAll(jsonFile)
			if err = json.Unmarshal(byteValue, &config); err != nil {
				log.Printf("error on unmarshal config from file %s\n", err.Error())
			} else {
				token = config.Token

				initFromFile = true
			}
		}
	}

	if !initFromFile {
		flag.StringVar(&token, "TOKEN", lookupEnvOrString("TOKEN", token), "telegram bot token")
		flag.Parse()
	}

	if token == "" {
		log.Fatal("TOKEN env var not set")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b := bot.New(token, opts...)

	log.Println("bot started")

	b.Start(ctx)
}

func lookupEnvOrString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultVal
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery != nil {
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}

		messageText := update.CallbackQuery.Message.Text
		if update.CallbackQuery.Message.ReplyMarkup != nil {
			switch c := update.CallbackQuery.Message.ReplyMarkup.(type) {
			case map[string]interface{}:
				buttons := []models.InlineKeyboardButton{}
				items := []string{}
				for _, v := range c["inline_keyboard"].([]interface{}) {
					subitems := v.([]interface{})
					for _, i := range subitems {
						text := i.(map[string]interface{})["text"].(string)
						callbackData := i.(map[string]interface{})["callback_data"].(string)

						if callbackData == update.CallbackQuery.Data {
							if strings.HasPrefix(callbackData, "busy-") {
								callbackData = strings.Replace(callbackData, "busy-", "free-", 1)
								text = strings.Replace(text, "🟢", "🏗️", 1)
							} else {
								callbackData = strings.Replace(callbackData, "free-", "busy-", 1)
								text = strings.Replace(text, "🏗️", "🟢", 1)
							}
						}

						items = append(items, text)

						buttons = append(buttons, models.InlineKeyboardButton{Text: text, CallbackData: callbackData})
					}
				}
				if len(items) > 0 {
					messageText = strings.Join(items, "  ")
				}

				kb.InlineKeyboard = [][]models.InlineKeyboardButton{buttons}
			default:
				// fmt.Printf("N%T\n", c)
			}
		}

		editedMessage := &bot.EditMessageTextParams{
			ChatID:      update.CallbackQuery.Message.Chat.ID,
			MessageID:   update.CallbackQuery.Message.ID,
			Text:        messageText,
			ReplyMarkup: kb,
		}

		b.EditMessageText(ctx, editedMessage)

		return
	}

	if update.Message != nil && update.Message.Text == "/start" {
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "🟢 testing-id", CallbackData: "busy-testing-id"},
					{Text: "🟢 testing-1", CallbackData: "busy-testing-1"},
					{Text: "🟢 testing-2", CallbackData: "busy-testing-2"},
				},
			},
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "🟢 testing-id  🟢 testing-1  🟢 testing-2",
			ReplyMarkup: kb,
		})

		return
	}

	log.Printf("message %#v\n", update)
}
