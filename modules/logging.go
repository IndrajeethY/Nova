package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetLogChat(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, err := eOR(m, locales.Tr("logging.usage_setlog"))
		return err
	}

	chatId, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, err = eOR(m, locales.Tr("logging.invalid_chat"))
		return err
	}

	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		_, err = eOR(m, locales.Tr("logging.assistant_not_in_chat"))
		return err
	}

	_, err = tgbot.SendMessage(peer, fmt.Sprintf(locales.Tr("logging.log_set_success"), args))
	if err != nil {
		_, err = eOR(m, locales.Tr("logging.send_error"))
		return err
	}

	db.Set("LOG_CHAT", strconv.FormatInt(chatId, 10))
	logger.Event("LOG_CHAT", fmt.Sprintf("Log channel set to %d", chatId))
	_, err = eOR(m, fmt.Sprintf(locales.Tr("logging.log_set_success"), args))
	return err
}

func GetLogChat(m *telegram.NewMessage) error {
	config := db.Get("LOG_CHAT")
	if config == "" {
		_, err := eOR(m, locales.Tr("logging.not_set"))
		return err
	}
	_, err := eOR(m, fmt.Sprintf(locales.Tr("logging.log_result"), config))
	return err
}

func DelLogChat(m *telegram.NewMessage) error {
	if !db.Exists("LOG_CHAT") {
		_, err := eOR(m, locales.Tr("logging.not_found"))
		return err
	}
	db.Del("LOG_CHAT")
	logger.SetLogToChannel(false)
	_, err := eOR(m, locales.Tr("logging.deleted"))
	return err
}

func ToggleLogging(m *telegram.NewMessage) error {
	args := strings.ToLower(m.Args())

	switch args {
	case "on", "enable":
		logger.SetLogToChannel(true)
		_, err := eOR(m, locales.Tr("logging.enabled"))
		return err
	case "off", "disable":
		logger.SetLogToChannel(false)
		_, err := eOR(m, locales.Tr("logging.disabled"))
		return err
	default:
		_, err := eOR(m, locales.Tr("logging.usage_toggle"))
		return err
	}
}

func LoadLoggingModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "setlog", Func: SetLogChat, Description: "Set log channel", ModuleName: "Logging"},
		{Command: "getlog", Func: GetLogChat, Description: "Get log channel", ModuleName: "Logging"},
		{Command: "dellog", Func: DelLogChat, Description: "Delete log channel", ModuleName: "Logging"},
		{Command: "logging", Func: ToggleLogging, Description: "Toggle logging on/off", ModuleName: "Logging"},
	}
	AddHandlers(handlers, c)
}
