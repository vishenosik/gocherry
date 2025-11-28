/*
package bot contains
*/
package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/vishenosik/gocherry/pkg/config"
	"github.com/vishenosik/gocherry/pkg/errors"
	"github.com/vishenosik/gocherry/pkg/logs"
)

var ErrCommandNotFound = errors.New("command not found")

type TelegramBot struct {
	bot  *tgbotapi.BotAPI
	conf Config

	commands map[string]HandlerFunc

	// sub

	log *slog.Logger

	// status

	running atomic.Bool
}

type Router interface {
	Route(tb *TelegramBot)
}

func NewBotAPIEnv(routers ...Router) (*TelegramBot, error) {
	var conf ConfigEnv
	err := config.ReadConfigEnv(&conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read env config")
	}
	return NewBotAPI(conf.ToConfig(), routers...)
}

func NewBotAPI(config Config, routers ...Router) (*TelegramBot, error) {

	if err := config.validate(); err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("NewBotAPI error: %w", err)
	}

	bot.Debug = true

	tb := &TelegramBot{
		bot:      bot,
		conf:     config,
		commands: make(map[string]HandlerFunc),
		log:      logs.SetupLogger().With(logs.AppComponent("telegram_bot")),
	}

	tb.log.Info("authorized on account", slog.String("username", bot.Self.UserName))

	tb.running.Store(false)

	for _, router := range routers {
		router.Route(tb)
	}

	return tb, nil
}

func (tb *TelegramBot) Start(ctx context.Context) error {

	_, ok := tb.commands["/save"]

	tb.log.Debug("commands list", slog.Any("cmds", tb.commandsList()), slog.Bool("save", ok))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = tb.conf.Timeout

	tb.running.Store(true)

	updates := tb.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			msg := update.Message

			if msg == nil {
				continue
			}

			handler, err := tb.handlerFromMsg(msg)
			if err != nil {
				tb.log.Error("command error", logs.Error(err))
				continue
			}

			handler(tb.bot, msg)
		}
	}
}

func (tb *TelegramBot) Stop(ctx context.Context) error {
	tb.running.Store(false)
	return nil
}

func (tb *TelegramBot) handlerFromMsg(msg *tgbotapi.Message) (HandlerFunc, error) {
	if msg.IsCommand() {
		command := "/" + msg.Command()
		handler, ok := tb.commands[command]
		if !ok {
			return nil, errors.Wrap(ErrCommandNotFound, command)
		}
		return handler, nil
	}

	if capt := msg.Caption; capt != "" {

		for _, command := range tb.commandsList() {
			if !strings.Contains(capt, command) {
				continue
			}
			handler, ok := tb.commands[command]
			if !ok {
				return nil, errors.Wrap(ErrCommandNotFound, command)
			}
			return handler, nil
		}
	}

	return nil, ErrCommandNotFound
}

func (tb *TelegramBot) commandsList() []string {
	commands := make([]string, 0, len(tb.commands))
	for command := range tb.commands {
		commands = append(commands, command)
	}
	return commands
}
