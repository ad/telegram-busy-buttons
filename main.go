package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const ConfigFileName = "/data/options.json"

// Config ...
type Config struct {
	Token string `json:"TOKEN"`
}

type CallbackData struct {
	Command string  `json:"c"`
	User    string  `json:"u,omitempty"`
	Notify  []int64 `json:"n,omitempty"`
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

	b, newBotError := bot.New(token, opts...)
	if newBotError != nil {
		log.Println("start bot error", newBotError)

		return
	}

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
		cbdMessage := &CallbackData{}

		isNotify := false

		target := ""
		if strings.HasPrefix(update.CallbackQuery.Data, "free-") || strings.HasPrefix(update.CallbackQuery.Data, "busy-") {
			isNotify = strings.HasPrefix(update.CallbackQuery.Data, "⚡")

			target = strings.TrimPrefix(strings.TrimPrefix(update.CallbackQuery.Data, "free-"), "busy-")
		} else {
			if err := json.Unmarshal([]byte(update.CallbackQuery.Data), cbdMessage); err != nil {
				log.Printf("error on unmarshal callback data %s\n", err.Error())
			} else {
				isNotify = strings.HasPrefix(cbdMessage.Command, "⚡")

				target = strings.TrimPrefix(strings.TrimPrefix(cbdMessage.Command, "free-"), "busy-")
			}
		}

		notificationText := fmt.Sprintf(
			"%s updated by %s %s",
			target,
			update.CallbackQuery.Sender.FirstName,
			update.CallbackQuery.Sender.LastName,
		)

		if isNotify {
			notificationText = fmt.Sprintf(
				"%s %s requested notifications",
				update.CallbackQuery.Sender.FirstName,
				update.CallbackQuery.Sender.LastName,
			)
		}

		log.Printf("%#v from %d\n", notificationText, update.CallbackQuery.Message.Chat.ID)

		// hide Loading... message and show who pressed button
		_, _ = b.AnswerCallbackQuery(
			ctx,
			&bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            notificationText,
			},
		)

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}

		messageText := update.CallbackQuery.Message.Text
		if update.CallbackQuery.Message.ReplyMarkup != nil {
			switch c := update.CallbackQuery.Message.ReplyMarkup.(type) {
			case map[string]interface{}:
				buttons := []models.InlineKeyboardButton{}
				items := []string{}

				notifyButtonPresent := &CallbackData{}
				notifyUsers := []int64{}
				for _, v := range c["inline_keyboard"].([]interface{}) {
					subitems := v.([]interface{})
					for _, i := range subitems {
						callbackData := i.(map[string]interface{})["callback_data"].(string)

						cbd := &CallbackData{}

						if err := json.Unmarshal([]byte(callbackData), cbd); err == nil {
							if strings.HasPrefix(cbd.Command, "⚡") {
								notifyButtonPresent = cbd
								foundInNotify := false

								if cbd.Notify != nil {
									for _, i := range cbd.Notify {
										if cbd.Command == cbdMessage.Command {
											if i != update.CallbackQuery.Sender.ID {
												notifyUsers = append(notifyUsers, i)
											} else {
												foundInNotify = true
											}
										} else {
											notifyUsers = append(notifyUsers, i)
										}
									}
								}

								if cbd.Command == cbdMessage.Command {
									if !foundInNotify {
										notifyUsers = append(notifyUsers, update.CallbackQuery.Sender.ID)
									}
								}

								cbd.Command = "⚡"
								cbd.Notify = notifyUsers

								if len(cbd.Notify) > 0 {
									cbd.Command = fmt.Sprintf("⚡%d", len(cbd.Notify))
								}
							}
						}
					}
				}

				for _, v := range c["inline_keyboard"].([]interface{}) {
					subitems := v.([]interface{})
					for _, i := range subitems {
						// fmt.Printf("%#v\n", i.(map[string]interface{}))
						callbackData := i.(map[string]interface{})["callback_data"].(string)
						text := i.(map[string]interface{})["text"].(string)

						cbd := &CallbackData{}

						if err := json.Unmarshal([]byte(callbackData), cbd); err != nil {
							log.Printf("error on unmarshal callback data %s\n", err.Error())
							if callbackData == update.CallbackQuery.Data {
								cbd.User = shortenUsername(callbackData, update.CallbackQuery.Sender.FirstName, update.CallbackQuery.Sender.LastName)

								if strings.HasPrefix(callbackData, "busy-") {
									text = strings.Replace(text, "🟢", "🏗️", 1)
									cbd.Command = strings.Replace(callbackData, "busy-", "free-", 1)
								} else if strings.HasPrefix(callbackData, "free-") {
									text = strings.Replace(text, "🏗️", "🟢", 1)
									cbd.Command = strings.Replace(callbackData, "free-", "busy-", 1)
								}
							} else {
								cbd.Command = callbackData
							}
						} else {
							if cbd.Command == cbdMessage.Command {
								if !strings.HasPrefix(cbd.Command, "⚡") {
									cbd.User = shortenUsername(cbd.Command, update.CallbackQuery.Sender.FirstName, update.CallbackQuery.Sender.LastName)

									if strings.HasPrefix(cbd.Command, "busy-") {
										text = strings.Replace(text, "🟢", "🏗️", 1)
										cbd.Command = strings.Replace(cbd.Command, "busy-", "free-", 1)
									} else if strings.HasPrefix(cbd.Command, "free-") {
										text = strings.Replace(text, "🏗️", "🟢", 1)
										cbd.Command = strings.Replace(cbd.Command, "free-", "busy-", 1)
									}
								}
							}
						}

						if !strings.HasPrefix(cbd.Command, "⚡") {
							itemText := text
							if strings.HasPrefix(cbd.Command, "free-") && cbd.User != "" {
								itemText = fmt.Sprintf("%s (%s)", text, cbd.User)
							}

							if !strings.HasPrefix(cbdMessage.Command, "⚡") && len(notifyUsers) > 0 {
								// notify users
								for _, userID := range notifyUsers {
									_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
										ChatID: userID,
										Text:   fmt.Sprintf("%s status updated", itemText),
									})
								}
							}

							items = append(items, itemText)

							cbdToSend, err := json.Marshal(cbd)
							if err != nil {
								log.Printf("%#v, err %s\n", cbd, err)

								return
							}

							buttons = append(
								buttons,
								models.InlineKeyboardButton{
									CallbackData: string(cbdToSend),
									Text:         text,
								},
							)
						}
					}
				}

				if notifyButtonPresent.Command == "" {
					notifyButtonPresent.Command = "⚡"
				}

				cbdToSend, err := json.Marshal(notifyButtonPresent)
				if err != nil {
					log.Printf("%#v, err %s\n", notifyButtonPresent, err)

					return
				}

				buttons = append(
					buttons,
					models.InlineKeyboardButton{
						CallbackData: string(cbdToSend),
						Text:         notifyButtonPresent.Command,
					},
				)

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

		_, err := b.EditMessageText(ctx, editedMessage)
		if err != nil {
			log.Printf("error on edit message %s, %#v %#v\n", err.Error(), editedMessage, editedMessage.ReplyMarkup)
		}

		return
	}

	if update.Message != nil && strings.HasPrefix(update.Message.Text, "/create") {
		message := strings.Trim(regexp.MustCompile(`\s+`).ReplaceAllString(update.Message.Text, " "), " ")
		parts := strings.Fields(message)

		log.Printf("message %#v from %d\n", message, update.Message.Chat.ID)

		if len(parts) < 2 {
			_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "you must send command in format /create name1 name2 nameN",
			})

			return
		}

		messageText := ""

		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{},
		}

		buttons := []models.InlineKeyboardButton{}
		items := []string{}
		for _, v := range parts[1:] {
			text := "🟢 " + v
			callbackData := CallbackData{
				Command: "busy-" + v,
			}

			cbd, err := json.Marshal(callbackData)
			if err != nil {
				_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   "Failed to create buttons",
				})

				return
			}

			items = append(items, text)
			buttons = append(
				buttons,
				models.InlineKeyboardButton{
					CallbackData: string(cbd),
					Text:         text,
				},
			)
		}
		// old style
		// for _, v := range parts[1:] {
		// 	text := "🟢 " + v
		// 	callbackData := CallbackData{
		// 		Command: "busy-" + v,
		// 	}

		// 	cbd, err := json.Marshal(callbackData)
		// 	if err != nil {
		// 		b.SendMessage(ctx, &bot.SendMessageParams{
		// 			ChatID: update.Message.Chat.ID,
		// 			Text:   "Failed to create buttons",
		// 		})

		// 		return
		// 	}

		// 	items = append(items, text)
		// 	buttons = append(
		// 		buttons,
		// 		models.InlineKeyboardButton{
		// 			CallbackData: string(cbd),
		// 			Text:         text,
		// 		},
		// 	)
		// }
		if len(items) > 0 {
			messageText = strings.Join(items, "  ")
		}

		notifyButtonPresent := &CallbackData{
			Command: "⚡",
		}

		cbdToSend, err := json.Marshal(notifyButtonPresent)
		if err != nil {
			log.Printf("%#v, err %s\n", notifyButtonPresent, err)

			return
		}

		buttons = append(
			buttons,
			models.InlineKeyboardButton{
				CallbackData: string(cbdToSend),
				Text:         notifyButtonPresent.Command,
			},
		)

		kb.InlineKeyboard = [][]models.InlineKeyboardButton{buttons}

		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        messageText,
			ReplyMarkup: kb,
		})

		return
	}
}

// shorten text on button to limit 64 chars
func shortenUsername(command, name, lastname string) string {
	// 18 chars is allocated to struct {"c": "", "u": ""}

	limit := 64 - 18 - len(command)

	if limit <= 0 {
		return ""
	}

	if len(name) <= 0 && len(lastname) <= 0 {
		return ""
	}

	if len(name)+len(lastname) < limit {
		return name + " " + lastname
	}

	if len(name) > 0 {
		if len(name) < limit-3 {
			return name + " " + string([]rune(lastname)[0:1]) + "."
		}

		if len(name) > limit {
			if len([]rune(name)) >= limit {
				return string([]rune(name)[0:limit])
			}
		}

		if len(lastname) > limit {
			return string([]rune(lastname)[0:limit])
		}
	}

	return ""
}
