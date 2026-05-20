package modules

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetTagLogger(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, err := m.Edit("<code>Usage: .taglogger &lt;chat_id&gt;</code>")
		return err
	}
	chatId, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, err = m.Edit("<code>Invalid chat id</code>")
		return err
	}
	peer, err := tgbot.GetSendablePeer(chatId)
	if err != nil {
		_, err = m.Edit("<code>Assistant is not in the chat</code>")
		return err
	}
	_, err = tgbot.SendMessage(peer, fmt.Sprintf("<b>Tag logger set successfully:</b>\n<b>Chat ID:</b> <code>%s</code>", args))
	if err != nil {
		_, err = m.Edit("<code>Error sending message to the chat</code>")
		return err
	}
	err = Db.Set(context.Background(), "TAG_LOGGER", chatId, 0).Err()
	if err != nil {
		_, err = m.Edit("<code>Error setting tag logger</code>")
		return err
	}
	_, err = m.Edit(fmt.Sprintf("<b>Tag logger set successfully:</b>\n<b>Chat ID:</b> <code>%s</code>", args))
	return err
}

func GetTagLogger(m *telegram.NewMessage) error {
	config, err := Db.Get(context.Background(), "TAG_LOGGER").Result()
	if err != nil {
		_, err = m.Edit("<code>Tag logger not set</code>")
		return err
	}
	_, err = m.Edit(fmt.Sprintf("<b>Tag logger:</b> <code>%s</code>", config))
	return err
}

func DelTagLogger(m *telegram.NewMessage) error {
	err := Db.Del(context.Background(), "TAG_LOGGER").Err()
	if err != nil {
		_, err = m.Edit("<code>Tag logger not found</code>")
		return err
	}
	_, err = m.Edit("<b>Tag logger deleted successfully</b>")
	return err
}

func CheckForTags(m *telegram.NewMessage) error {
	if m.Message.Mentioned {
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
			log.Println("Error getting sendable peer:", err)
			return err
		}
		btn := telegram.ButtonBuilder{}
		var msg string
		if m.Text() == "" {
			msg = "No Text or Might be a media"
		} else {
			msg = m.Text()
		}
		_, err = tgbot.SendMessage(peer, fmt.Sprintf("<b>Yᴏᴜ Hᴀᴠᴇ Bᴇᴇɴ Tᴀɢɢᴇᴅ:</b>\n\n<b>Cʜᴀᴛ Tɪᴛʟᴇ:</b> <code>%s</code>\n<b>Tᴀɢɢᴇᴅ Bʏ:</b> <code>%s</code>\n<b>Mᴇꜱꜱᴀɢᴇ:</b> <code>%s</code>", m.Channel.Title, m.Sender.FirstName, msg), &telegram.SendOptions{ParseMode: "HTML", ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL("🔗 Go to message", msgLink(m))).Build()})
		if err != nil {
			log.Println("Error sending message to chat:", err)
			return err
		}
		return nil
	}
	return nil
}

func LoadTagLogger(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "taglogger", Func: SetTagLogger, Description: "Set a chat as tag logger", ModuleName: "Tag Logger"},
		{Command: "gettaglogger", Func: GetTagLogger, Description: "Get the tag logger chat", ModuleName: "Tag Logger"},
		{Command: "deltaglogger", Func: DelTagLogger, Description: "Delete the tag logger chat", ModuleName: "Tag Logger"},
	}
	AddHandlers(handlers, c)
	c.AddMessageHandler(telegram.OnNewMessage, CheckForTags)
}
