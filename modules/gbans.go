package modules

import (
	"NovaUserbot/locales"
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

func init() {
	RegisterModule("Gban", loadGbanModule)
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

	currentBansJson, _ := Db.Get(context.Background(), "GBANS").Result()
	banMap := make(map[int64]BanInfo)
	if currentBansJson != "" {
		json.Unmarshal([]byte(currentBansJson), &banMap)
	}

	if banInfo, exists := banMap[userID]; exists {
		_, err := eOR(m, locales.Trf("gban.already_banned", userID, Name, banInfo.Reason, banInfo.Time))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.banning"))

	banMap[userID] = BanInfo{
		Reason: reason,
		Time:   time.Now().Format(time.RFC1123),
	}

	updatedBansJson, _ := json.Marshal(banMap)
	err := Db.Set(context.Background(), "GBANS", updatedBansJson, 0).Err()
	if err != nil {
		_, err := msg.Edit(locales.Tr("gban.ban_error"))
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

	logMessage(locales.Trf("gban.log_banned", userID, Name, reason))

	_, err = msg.Edit(locales.Trf("gban.banned", userID, Name, reason, success))
	return err
}

func ungbanUser(m *telegram.NewMessage) error {
	userID, Name, _ := ExtractUserMsg(m)
	if userID == 0 {
		_, err := eOR(m, locales.Tr("gban.usage_ungban"))
		return err
	}

	currentBansJson, _ := Db.Get(context.Background(), "GBANS").Result()
	banMap := make(map[int64]BanInfo)
	if currentBansJson != "" {
		json.Unmarshal([]byte(currentBansJson), &banMap)
	}

	if _, exists := banMap[userID]; !exists {
		_, err := eOR(m, locales.Trf("gban.not_banned", userID, Name))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.unbanning"))

	delete(banMap, userID)

	updatedBansJson, _ := json.Marshal(banMap)
	err := Db.Set(context.Background(), "GBANS", updatedBansJson, 0).Err()
	if err != nil {
		_, err := msg.Edit(locales.Tr("gban.unban_error"))
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

	logMessage(locales.Trf("gban.log_unbanned", userID, Name))

	_, err = msg.Edit(locales.Trf("gban.unbanned", userID, Name, success))
	return err
}

func gbanned(m *telegram.NewMessage) error {
	currentBansJson, err := Db.Get(context.Background(), "GBANS").Result()
	if err != nil || currentBansJson == "" {
		_, err := eOR(m, locales.Tr("gban.list_empty"))
		return err
	}

	banMap := make(map[int64]BanInfo)
	json.Unmarshal([]byte(currentBansJson), &banMap)

	msg, _ := eOR(m, locales.Tr("gban.fetching"))
	response := locales.Tr("gban.list_header")
	for userID, banInfo := range banMap {
		response += fmt.Sprintf(locales.Tr("gban.list_entry"), userID, banInfo.Reason)
	}
	_, err = msg.Edit(response)
	return err
}

func toggleAntispam(m *telegram.NewMessage) error {
	args := strings.ToLower(m.Args())
	if args == "" {
		if Db.SIsMember(context.Background(), "ANTISPAM_DISABLED", m.Chat.ID).Val() {
			_, err := eOR(m, locales.Tr("gban.antispam_off"))
			return err
		}
		_, err := eOR(m, locales.Tr("gban.antispam_on"))
		return err
	}
	if args == "enable" {
		err := Db.SRem(context.Background(), "ANTISPAM_DISABLED", m.Chat.ID).Err()
		if err != nil {
			log.Error("Error enabling antispam:", err)
			return err
		}
		_, err = eOR(m, locales.Tr("gban.antispam_enabled"))
		return err
	} else if args == "disable" {
		err := Db.SAdd(context.Background(), "ANTISPAM_DISABLED", m.Chat.ID).Err()
		if err != nil {
			log.Error("Error disabling antispam:", err)
			return err
		}
		_, err = eOR(m, locales.Tr("gban.antispam_disabled"))
		return err
	}
	_, err := eOR(m, locales.Tr("gban.antispam_usage"))
	return err
}

func loadGbanModule() {
	handlers := []*Handler{
		{ModuleName: "Gban", Command: "gban", Description: "Globally ban a user", Func: gbanUser},
		{ModuleName: "Gban", Command: "ungban", Description: "Globally unban a user", Func: ungbanUser},
		{ModuleName: "Gban", Command: "antispam", Description: "Toggle antispam in a chat", Func: toggleAntispam},
		{ModuleName: "Gban", Command: "gbanned", Description: "List all globally banned users", Func: gbanned},
	}
	AddHandlers(handlers, client)
}
