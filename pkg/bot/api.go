package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(bot *tgbotapi.BotAPI, msg *tgbotapi.Message)

func (tb *TelegramBot) Route(command string, handler HandlerFunc) {

	if tb.running.Load() {
		panic("adding command after agent started is not possible")
	}

	if command == "" {
		panic("command already added")
	}

	command = "/" + command

	if _, ok := tb.commands[command]; ok {
		panic(fmt.Sprintf("command %s already added", command))
	}

	tb.commands[command] = handler

}

func (tb *TelegramBot) Bot() *tgbotapi.BotAPI {
	return tb.bot
}
