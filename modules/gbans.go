package modules

import (
	"NovaUserbot/locales"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/go-redis/redis/v8"
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

	ctx := context.Background()
	var alreadyBanned bool
	var existingInfo BanInfo

	err := Db.Watch(ctx, func(tx *redis.Tx) error {
		bansJson, _ := tx.Get(ctx, "GBANS").Result()
		banMap := make(map[int64]BanInfo)
		if bansJson != "" {
			if err := json.Unmarshal([]byte(bansJson), &banMap); err != nil {
				return err
			}
		}

		if info, exists := banMap[userID]; exists {
			alreadyBanned = true
			existingInfo = info
			return nil
		}

		banMap[userID] = BanInfo{
			Reason: reason,
			Time:   time.Now().Format(time.RFC1123),
		}

		updated, err := json.Marshal(banMap)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "GBANS", updated, 0)
			return nil
		})
		return err
	}, "GBANS")

	if err != nil {
		_, err := eOR(m, locales.Tr("gban.ban_error"))
		return err
	}

	if alreadyBanned {
		_, err := eOR(m, locales.Trf("gban.already_banned", userID, Name, existingInfo.Reason, existingInfo.Time))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.banning"))

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

	ctx := context.Background()
	var notBanned bool

	err := Db.Watch(ctx, func(tx *redis.Tx) error {
		bansJson, _ := tx.Get(ctx, "GBANS").Result()
		banMap := make(map[int64]BanInfo)
		if bansJson != "" {
			if err := json.Unmarshal([]byte(bansJson), &banMap); err != nil {
				return err
			}
		}

		if _, exists := banMap[userID]; !exists {
			notBanned = true
			return nil
		}

		delete(banMap, userID)

		updated, err := json.Marshal(banMap)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "GBANS", updated, 0)
			return nil
		})
		return err
	}, "GBANS")

	if err != nil {
		_, err := eOR(m, locales.Tr("gban.unban_error"))
		return err
	}

	if notBanned {
		_, err := eOR(m, locales.Trf("gban.not_banned", userID, Name))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.unbanning"))

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
	if err := json.Unmarshal([]byte(currentBansJson), &banMap); err != nil {
		log.Error("Error unmarshaling GBANS:", err)
		_, err := eOR(m, locales.Tr("gban.list_empty"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("gban.fetching"))
	var sb strings.Builder
	sb.WriteString(locales.Tr("gban.list_header"))
	for userID, banInfo := range banMap {
		sb.WriteString(fmt.Sprintf(locales.Tr("gban.list_entry"), userID, banInfo.Reason))
	}
	_, err = msg.Edit(sb.String())
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
	switch args {
	case "enable":
		err := Db.SRem(context.Background(), "ANTISPAM_DISABLED", m.Chat.ID).Err()
		if err != nil {
			log.Error("Error enabling antispam:", err)
			return err
		}
		_, err = eOR(m, locales.Tr("gban.antispam_enabled"))
		return err
	case "disable":
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
		{ModuleName: "Gban", Command: "gban", Description: locales.Tr("desc.gban"), Func: gbanUser},
		{ModuleName: "Gban", Command: "ungban", Description: locales.Tr("desc.ungban"), Func: ungbanUser},
		{ModuleName: "Gban", Command: "antispam", Description: locales.Tr("desc.antispam"), Func: toggleAntispam},
		{ModuleName: "Gban", Command: "gbanned", Description: locales.Tr("desc.gbanned"), Func: gbanned},
	}
	AddHandlers(handlers, client)
}
