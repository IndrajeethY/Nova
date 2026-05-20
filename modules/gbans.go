package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	log "github.com/sirupsen/logrus"
)

type BanInfo struct {
	Reason string
	Time   string
}

func gbanUser(m *telegram.NewMessage) error {
	userID, Name, reason := ExtractUserMsg(m)
	if userID == 0 {
		_, err := eOR(m, "<code>Usage: .gban <user_id> or reply to a user</code>")
		return err
	}
	if userID == ubId {
		_, err := eOR(m, "<code>Can't globally ban myself</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}

	currentBansJson, _ := Db.Get(context.Background(), "GBANS").Result()
	banMap := make(map[int64]BanInfo)
	if currentBansJson != "" {
		json.Unmarshal([]byte(currentBansJson), &banMap)
	}

	if banInfo, exists := banMap[userID]; exists {
		_, err := eOR(m, fmt.Sprintf("<b>User <a href='tg://user?id=%d'>%s</a> is already globally banned\nReason:</b> %s\n<b>Time:</b> %s", userID, Name, banInfo.Reason, banInfo.Time))
		return err
	}

	msg, _ := eOR(m, "<code>Globally banning user...</code>")

	banMap[userID] = BanInfo{
		Reason: reason,
		Time:   time.Now().Format(time.RFC1123),
	}

	updatedBansJson, _ := json.Marshal(banMap)
	err := Db.Set(context.Background(), "GBANS", updatedBansJson, 0).Err()
	if err != nil {
		_, err := msg.Edit("<code>Error globally banning user</code>")
		return err
	}

	chats := m.Client.Cache.InputPeers.InputChannels
	var success int
	for chatID, chatHash := range chats {
		_, err := m.Client.EditBanned(&telegram.InputPeerChannel{ChannelID: chatID, AccessHash: chatHash}, userID, &telegram.BannedOptions{Ban: true})
		if err != nil {
			continue
		}
		success++
	}

	if err := logMessage(fmt.Sprintf("<b>#Globally Banned</b>\n<b>User:</b> <a href='tg://user?id=%d'>%s</a>\n<b>Reason:</b> %s", userID, Name, reason)); err != nil {
		return err
	}

	_, err = msg.Edit(fmt.Sprintf("<b>Globally banned <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s\n<b>Success:</b> %d", userID, Name, reason, success))
	return err
}

func ungbanUser(m *telegram.NewMessage) error {
	userID, Name, _ := ExtractUserMsg(m)
	if userID == 0 {
		_, err := eOR(m, "<code>Usage: .ungban <user_id> or reply to a user</code>")
		return err
	}

	currentBansJson, _ := Db.Get(context.Background(), "GBANS").Result()
	banMap := make(map[int64]BanInfo)
	if currentBansJson != "" {
		json.Unmarshal([]byte(currentBansJson), &banMap)
	}

	if _, exists := banMap[userID]; !exists {
		_, err := eOR(m, fmt.Sprintf("<b>User <a href='tg://user?id=%d'>%s</a> is not globally banned</b>", userID, Name))
		return err
	}

	msg, _ := eOR(m, "<code>Unglobally banning user...</code>")

	delete(banMap, userID)

	updatedBansJson, _ := json.Marshal(banMap)
	err := Db.Set(context.Background(), "GBANS", updatedBansJson, 0).Err()
	if err != nil {
		_, err := msg.Edit("<code>Error removing global ban</code>")
		return err
	}

	chats := m.Client.Cache.InputPeers.InputChannels
	var success int
	for chatID, chatHash := range chats {
		_, err := m.Client.EditBanned(&telegram.InputPeerChannel{ChannelID: chatID, AccessHash: chatHash}, userID, &telegram.BannedOptions{Unban: true})
		if err != nil {
			continue
		}
		success++
	}

	if err := logMessage(fmt.Sprintf("<b>#Globally Unbanned</b>\n<b>User:</b> <a href='tg://user?id=%d'>%s</a>", userID, Name)); err != nil {
		return err
	}

	_, err = msg.Edit(fmt.Sprintf("<b>Unglobally banned <a href='tg://user?id=%d'>%s</a></b>\n<b>Success:</b> %d", userID, Name, success))
	return err
}

func gbanned(m *telegram.NewMessage) error {
	currentBansJson, err := Db.Get(context.Background(), "GBANS").Result()
	if err != nil || currentBansJson == "" {
		_, err := eOR(m, "<b>No users are globally banned</b>")
		return err
	}

	banMap := make(map[int64]BanInfo)
	json.Unmarshal([]byte(currentBansJson), &banMap)

	msg, _ := eOR(m, "<code>Fetching globally banned users...</code>")
	response := "<b>Globally banned users:</b>\n"
	for userID, banInfo := range banMap {
		response += fmt.Sprintf("<b>▸</b> <code>%d</code> <b>Reason:</b> %s\n\n", userID, banInfo.Reason)
	}
	_, err = msg.Edit(response)
	return err
}

func toggleAntispam(m *telegram.NewMessage) error {
	args := strings.ToLower(m.Args())
	if args == "" {
		if Db.SIsMember(context.Background(), "ANTISPAM", m.Chat.ID).Val() {
			_, err := eOR(m, "<b>Antispam is disabled</b>")
			return err
		} else {
			_, err := eOR(m, "<b>Antispam is enabled</b>")
			return err
		}
	}
	if args == "enable" {
		err := Db.SRem(context.Background(), "ANTISPAM", m.Chat.ID).Err()
		if err != nil {
			log.Error("Error enabling antispam:", err)
			return err
		}
		_, err = m.Reply("Antispam has been enabled.")
		return err
	} else if args == "disable" {
		err := Db.SAdd(context.Background(), "ANTISPAM", m.Chat.ID).Err()
		if err != nil {
			log.Error("Error disabling antispam:", err)
			return err
		}
		_, err = m.Reply("Antispam has been disabled.")
		return err
	} else {
		_, err := m.Reply("Usage: .antispam <enable|disable>")
		return err
	}
}

func LoadGbanHandler(c *telegram.Client) {
	handlers := []*Handler{
		{
			ModuleName:  "Gban",
			Command:     "gban",
			Description: "Globally ban a user",
			Func:        gbanUser,
		},
		{
			ModuleName:  "Gban",
			Command:     "ungban",
			Description: "Globally unban a user",
			Func:        ungbanUser,
		},
		{
			ModuleName:  "Gban",
			Command:     "antispam",
			Description: "Enable or disable antispam in a chat",
			Func:        toggleAntispam,
		},
		{
			ModuleName:  "Gban",
			Command:     "gbanned",
			Description: "List all globally banned users",
			Func:        gbanned,
		},
	}
	AddHandlers(handlers, c)
}
