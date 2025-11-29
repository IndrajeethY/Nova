package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetTagLogger(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, err := m.Edit(locales.Tr("tag_logger.usage"))
		return err
	}

	chatId, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, err = m.Edit(locales.Tr("tag_logger.invalid_chat"))
		return err
	}

	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		_, err = m.Edit(locales.Tr("tag_logger.assistant_not_in_chat"))
		return err
	}

	_, err = tgbot.SendMessage(peer, fmt.Sprintf(locales.Tr("tag_logger.set_success"), args))
	if err != nil {
		_, err = m.Edit(locales.Tr("tag_logger.send_error"))
		return err
	}

	db.Set("TAG_LOGGER", strconv.FormatInt(chatId, 10))
	_, err = m.Edit(fmt.Sprintf(locales.Tr("tag_logger.set_success"), args))
	return err
}

func GetTagLogger(m *telegram.NewMessage) error {
	config := db.Get("TAG_LOGGER")
	if config == "" {
		_, err := m.Edit(locales.Tr("tag_logger.not_set"))
		return err
	}
	_, err := m.Edit(fmt.Sprintf(locales.Tr("tag_logger.get_result"), config))
	return err
}

func DelTagLogger(m *telegram.NewMessage) error {
	if !db.Exists("TAG_LOGGER") {
		_, err := m.Edit(locales.Tr("tag_logger.not_found"))
		return err
	}
	db.Del("TAG_LOGGER")
	_, err := m.Edit(locales.Tr("tag_logger.deleted"))
	return err
}

func CheckForTags(m *telegram.NewMessage) error {
	if !m.Message.Mentioned || m.Message.Out || m.Sender.Bot {
		return nil
	}

	config := db.Get("TAG_LOGGER")
	if config == "" {
		return nil
	}

	chatId := utils.StringToInt64(config)
	if m.ChatID() == chatId {
		return nil
	}

	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		return err
	}

	msgText := m.Text()
	if msgText == "" {
		msgText = locales.Tr("tag_logger.no_text")
	}

	btn := telegram.ButtonBuilder{}
	notification := fmt.Sprintf(locales.Tr("tag_logger.notification"), m.Channel.Title, m.Sender.FirstName, msgText)

	_, err = tgbot.SendMessage(peer, notification, &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL(locales.Tr("tag_logger.go_to_message"), msgLink(m))).Build(),
	})
	return err
}

func LoadTagLogger(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "taglogger", Func: SetTagLogger, Description: "Set tag logger chat", ModuleName: "Tag Logger"},
		{Command: "gettaglogger", Func: GetTagLogger, Description: "Get tag logger chat", ModuleName: "Tag Logger"},
		{Command: "deltaglogger", Func: DelTagLogger, Description: "Delete tag logger", ModuleName: "Tag Logger"},
	}
	AddHandlers(handlers, c)
	c.AddMessageHandler(telegram.OnNewMessage, CheckForTags)
}
