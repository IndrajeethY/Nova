package modules

import (
	"NovaUserbot/locales"
	"context"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Tag Logger", loadTagLoggerModule)
}

func SetTagLogger(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, err := eOR(m, locales.Tr("tag_logger.usage"))
		return err
	}
	chatId, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.invalid_chat"))
		return err
	}
	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.assistant_not_in_chat"))
		return err
	}
	_, err = tgbot.SendMessage(peer, locales.Trf("tag_logger.set_success", args))
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.send_error"))
		return err
	}
	err = Db.Set(context.Background(), "TAG_LOGGER", chatId, 0).Err()
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.set_error"))
		return err
	}
	_, err = eOR(m, locales.Trf("tag_logger.set_success", args))
	return err
}

func GetTagLogger(m *telegram.NewMessage) error {
	val, err := Db.Get(context.Background(), "TAG_LOGGER").Result()
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.not_set"))
		return err
	}
	_, err = eOR(m, fmt.Sprintf(locales.Tr("tag_logger.get_result"), val))
	return err
}

func DelTagLogger(m *telegram.NewMessage) error {
	err := Db.Del(context.Background(), "TAG_LOGGER").Err()
	if err != nil {
		_, err = eOR(m, locales.Tr("tag_logger.not_found"))
		return err
	}
	_, err = eOR(m, locales.Tr("tag_logger.deleted"))
	return err
}

func CheckForTags(m *telegram.NewMessage) error {
	if !m.Message.Mentioned {
		return nil
	}
	config, err := Db.Get(context.Background(), "TAG_LOGGER").Result()
	if err != nil {
		return nil
	}
	chatId, err := strconv.ParseInt(config, 10, 64)
	if err != nil {
		log.Println("Error parsing chat id:", err)
		return err
	}
	if m.ChatID() == chatId {
		return nil
	}
	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		return err
	}
	btn := telegram.ButtonBuilder{}
	msg := m.Text()
	if msg == "" {
		msg = locales.Tr("tag_logger.no_text")
	}
	_, err = tgbot.SendMessage(peer,
		locales.Trf("tag_logger.notification", m.Channel.Title, m.Sender.FirstName, msg),
		&telegram.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL("🔗 Go to message", msgLink(m))).Build(),
		},
	)
	return err
}

func loadTagLoggerModule() {
	handlers := []*Handler{
		{Command: "taglogger", Func: SetTagLogger, Description: "Set a chat as tag logger", ModuleName: "Tag Logger"},
		{Command: "gettaglogger", Func: GetTagLogger, Description: "Get the tag logger chat", ModuleName: "Tag Logger"},
		{Command: "deltaglogger", Func: DelTagLogger, Description: "Delete the tag logger chat", ModuleName: "Tag Logger"},
	}
	AddHandlers(handlers, client)
	client.AddMessageHandler(telegram.OnNewMessage, CheckForTags)
}
