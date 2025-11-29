package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type BanInfo struct {
	Reason string
	Time   string
}

func gbanUser(m *telegram.NewMessage) error {
	userID, Name, reason := ExtractUserMsg(m)
	if userID == 0 {
		_, err := eOR(m, locales.Tr("gban.usage_gban"))
		return err
	}

	if userID == ubId {
		_, err := eOR(m, locales.Tr("gban.cant_ban_self"))
		return err
	}

	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}

	banMap := make(map[int64]BanInfo)
	if data := db.Get("GBANS"); data != "" {
		json.Unmarshal([]byte(data), &banMap)
	}

	if info, exists := banMap[userID]; exists {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("gban.already_banned"), userID, Name, info.Reason, info.Time))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.banning"))

	banMap[userID] = BanInfo{Reason: reason, Time: time.Now().Format(time.RFC1123)}
	data, _ := json.Marshal(banMap)
	db.Set("GBANS", string(data))

	chats := m.Client.Cache.InputPeers.InputChannels
	var success int
	for chatID, chatHash := range chats {
		_, err := m.Client.EditBanned(&telegram.InputPeerChannel{ChannelID: chatID, AccessHash: chatHash}, userID, &telegram.BannedOptions{Ban: true})
		if err == nil {
			success++
		}
	}

	logMessage(fmt.Sprintf(locales.Tr("gban.log_banned"), userID, Name, reason))
	_, err := msg.Edit(fmt.Sprintf(locales.Tr("gban.banned"), userID, Name, reason, success))
	return err
}

func ungbanUser(m *telegram.NewMessage) error {
	userID, Name, _ := ExtractUserMsg(m)
	if userID == 0 {
		_, err := eOR(m, locales.Tr("gban.usage_ungban"))
		return err
	}

	banMap := make(map[int64]BanInfo)
	if data := db.Get("GBANS"); data != "" {
		json.Unmarshal([]byte(data), &banMap)
	}

	if _, exists := banMap[userID]; !exists {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("gban.not_banned"), userID, Name))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.unbanning"))

	delete(banMap, userID)
	data, _ := json.Marshal(banMap)
	db.Set("GBANS", string(data))

	chats := m.Client.Cache.InputPeers.InputChannels
	var success int
	for chatID, chatHash := range chats {
		_, err := m.Client.EditBanned(&telegram.InputPeerChannel{ChannelID: chatID, AccessHash: chatHash}, userID, &telegram.BannedOptions{Unban: true})
		if err == nil {
			success++
		}
	}

	logMessage(fmt.Sprintf(locales.Tr("gban.log_unbanned"), userID, Name))
	_, err := msg.Edit(fmt.Sprintf(locales.Tr("gban.unbanned"), userID, Name, success))
	return err
}

func gbanned(m *telegram.NewMessage) error {
	data := db.Get("GBANS")
	if data == "" {
		_, err := eOR(m, locales.Tr("gban.list_empty"))
		return err
	}

	banMap := make(map[int64]BanInfo)
	json.Unmarshal([]byte(data), &banMap)

	msg, _ := eOR(m, locales.Tr("gban.fetching"))

	response := locales.Tr("gban.list_header") + "\n"
	for userID, info := range banMap {
		response += fmt.Sprintf(locales.Tr("gban.list_entry"), userID, info.Reason) + "\n\n"
	}

	_, err := msg.Edit(response)
	return err
}

func toggleAntispam(m *telegram.NewMessage) error {
	args := strings.ToLower(m.Args())

	if args == "" {
		if db.SIsMember("ANTISPAM", m.Chat.ID) {
			_, err := eOR(m, locales.Tr("gban.antispam_off"))
			return err
		}
		_, err := eOR(m, locales.Tr("gban.antispam_on"))
		return err
	}

	switch args {
	case "enable":
		db.SRem("ANTISPAM", m.Chat.ID)
		_, err := m.Reply(locales.Tr("gban.antispam_enabled"))
		return err
	case "disable":
		db.SAdd("ANTISPAM", m.Chat.ID)
		_, err := m.Reply(locales.Tr("gban.antispam_disabled"))
		return err
	default:
		_, err := m.Reply(locales.Tr("gban.antispam_usage"))
		return err
	}
}

func LoadGbanHandler(c *telegram.Client) {
	handlers := []*Handler{
		{ModuleName: "Gban", Command: "gban", Description: "Globally ban a user", Func: gbanUser},
		{ModuleName: "Gban", Command: "ungban", Description: "Globally unban a user", Func: ungbanUser},
		{ModuleName: "Gban", Command: "antispam", Description: "Toggle antispam", Func: toggleAntispam},
		{ModuleName: "Gban", Command: "gbanned", Description: "List globally banned users", Func: gbanned},
	}
	AddHandlers(handlers, c)
}
