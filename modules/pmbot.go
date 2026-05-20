package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	pmBotEnabledKey  = "PM_BOT_ENABLED"
	pmBotLogGroupKey = "PM_BOT_LOG_GROUP"
	pmBotTopicPrefix = "PM_TOPIC:"
	pmBotBlockedKey  = "PM_BOT_BLOCKED"
)

func init() {
	RegisterModule("PM Bot", loadPmBotModule)
}

func isPmBotEnabled() bool {
	return Db.Get(context.Background(), pmBotEnabledKey).Val() == "true"
}

func getLogGroup() int64 {
	val := Db.Get(context.Background(), pmBotLogGroupKey).Val()
	if val == "" {
		return 0
	}
	return utils.StringToInt64(val)
}

func getTopicForUser(userID int64) int32 {
	val := Db.Get(context.Background(), pmBotTopicPrefix+strconv.FormatInt(userID, 10)).Val()
	if val == "" {
		return 0
	}
	id, _ := strconv.ParseInt(val, 10, 32)
	return int32(id)
}

func setTopicForUser(userID int64, topicID int32) {
	Db.Set(context.Background(), pmBotTopicPrefix+strconv.FormatInt(userID, 10), topicID, 0)
	Db.Set(context.Background(), fmt.Sprintf("PM_TOPIC_REVERSE:%d", topicID), strconv.FormatInt(userID, 10), 0)
}

func isUserBlocked(userID int64) bool {
	return Db.SIsMember(context.Background(), pmBotBlockedKey, userID).Val()
}

func blockUser(userID int64) {
	Db.SAdd(context.Background(), pmBotBlockedKey, userID)
}

func unblockUser(userID int64) {
	Db.SRem(context.Background(), pmBotBlockedKey, userID)
}

func resolveTopicUser(topicID int32) int64 {
	val := Db.Get(context.Background(), fmt.Sprintf("PM_TOPIC_REVERSE:%d", topicID)).Val()
	if val != "" {
		return utils.StringToInt64(val)
	}
	keys, _ := Db.Keys(context.Background(), pmBotTopicPrefix+"*").Result()
	for _, key := range keys {
		v := Db.Get(context.Background(), key).Val()
		tid, _ := strconv.ParseInt(v, 10, 32)
		if int32(tid) == topicID {
			uid := strings.TrimPrefix(key, pmBotTopicPrefix)
			Db.Set(context.Background(), fmt.Sprintf("PM_TOPIC_REVERSE:%d", topicID), uid, 0)
			return utils.StringToInt64(uid)
		}
	}
	return 0
}

func createTopicForUser(logGroup int64, user *telegram.UserObj) (int32, error) {
	peer, err := tgbot.GetSendablePeer(logGroup)
	if err != nil {
		return 0, err
	}

	title := user.FirstName
	if user.LastName != "" {
		title += " " + user.LastName
	}
	if len(title) > 128 {
		title = title[:128]
	}

	updates, err := tgbot.MessagesCreateForumTopic(&telegram.MessagesCreateForumTopicParams{
		Peer:     peer,
		Title:    title,
		RandomID: rand.Int63(),
	})
	if err != nil {
		return 0, err
	}

	updatesObj, ok := updates.(*telegram.UpdatesObj)
	if !ok {
		return 0, fmt.Errorf("unexpected updates type")
	}
	for _, upd := range updatesObj.Updates {
		if msgUpd, ok := upd.(*telegram.UpdateNewChannelMessage); ok {
			if svc, ok := msgUpd.Message.(*telegram.MessageService); ok {
				if _, ok := svc.Action.(*telegram.MessageActionTopicCreate); ok {
					return svc.ID, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("topic ID not found in updates")
}

func sendUserInfoToTopic(logGroup int64, topicID int32, user *telegram.UserObj) {
	username := "@" + user.Username
	if user.Username == "" {
		username = "N/A"
	}
	info := locales.Trf("pmbot.user_info",
		user.FirstName+" "+user.LastName,
		user.ID,
		username,
	)

	peer, _ := tgbot.GetSendablePeer(logGroup)

	photos, err := client.PhotosGetUserPhotos(
		&telegram.InputUserObj{UserID: user.ID, AccessHash: user.AccessHash},
		0, 0, 1,
	)
	if err == nil {
		if p, ok := photos.(*telegram.PhotosPhotosObj); ok && len(p.Photos) > 0 {
			if photo, ok := p.Photos[0].(*telegram.PhotoObj); ok {
				file, dlErr := client.DownloadMedia(photo)
				if dlErr == nil {
					tgbot.SendMessage(peer, info, &telegram.SendOptions{
						ParseMode: "HTML",
						Media:     file,
						TopicID:   topicID,
					})
					return
				}
			}
		}
	}

	tgbot.SendMessage(peer, info, &telegram.SendOptions{
		ParseMode: "HTML",
		TopicID:   topicID,
	})
}

func buildStartKeyboard() *telegram.ReplyInlineMarkup {
	btn := telegram.ButtonBuilder{}
	ownerUser, _ := tgbot.GetUser(ubId)
	if ownerUser == nil {
		return nil
	}
	return telegram.NewKeyboard().AddRow(
		btn.Mention("👤 "+locales.Tr("pmbot.owner_btn"), &telegram.InputUserObj{
			UserID:     ownerUser.ID,
			AccessHash: ownerUser.AccessHash,
		}),
	).Build()
}

func onBotPrivateMessage(m *telegram.NewMessage) error {
	if !m.IsPrivate() || m.Sender == nil || m.Sender.Bot {
		return nil
	}

	if m.Sender.ID == ubId || IsSudoer(m.Sender.ID) {
		return nil
	}

	if m.Text() == "/start" {
		return nil
	}

	if !isPmBotEnabled() {
		return nil
	}

	if isUserBlocked(m.Sender.ID) {
		_, err := m.Reply(locales.Tr("pmbot.blocked_msg"))
		return err
	}

	logGroup := getLogGroup()
	if logGroup == 0 {
		return nil
	}

	topicID := getTopicForUser(m.Sender.ID)
	if topicID == 0 {
		user, err := tgbot.GetUser(m.Sender.ID)
		if err != nil {
			return err
		}
		topicID, err = createTopicForUser(logGroup, user)
		if err != nil {
			log.Errorf("Failed to create topic for user %d: %v", m.Sender.ID, err)
			return err
		}
		setTopicForUser(m.Sender.ID, topicID)
		sendUserInfoToTopic(logGroup, topicID, user)
	}

	peer, err := tgbot.GetSendablePeer(logGroup)
	if err != nil {
		return err
	}

	opts := &telegram.SendOptions{TopicID: topicID}
	if m.Media() != nil {
		opts.Media = m.Media()
	}
	_, err = tgbot.SendMessage(peer, m.Text(), opts)
	return err
}

func onBotStart(m *telegram.NewMessage) error {
	if !m.IsPrivate() || m.Sender == nil || m.Sender.Bot {
		return nil
	}

	if m.Sender.ID == ubId || IsSudoer(m.Sender.ID) {
		return nil
	}

	ownerName := client.Me().FirstName
	if client.Me().LastName != "" {
		ownerName += " " + client.Me().LastName
	}

	text := locales.Trf("pmbot.start_msg", ownerName)
	if isPmBotEnabled() {
		text += locales.Tr("pmbot.start_msg_active")
	}

	markup := buildStartKeyboard()
	_, err := m.Reply(text, &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: markup,
	})
	return err
}

func onLogGroupReply(m *telegram.NewMessage) error {
	logGroup := getLogGroup()
	if logGroup == 0 || m.ChatID() != logGroup {
		return nil
	}

	if m.Sender != nil && m.Sender.ID == tbotId {
		return nil
	}

	topicID, isTopic := m.TopicID()
	if !isTopic || topicID == 0 || topicID == 1 {
		return nil
	}

	userID := resolveTopicUser(topicID)
	if userID == 0 {
		return nil
	}

	peer, err := tgbot.GetSendablePeer(userID)
	if err != nil {
		return err
	}

	opts := &telegram.SendOptions{}
	if m.Media() != nil {
		opts.Media = m.Media()
	}
	if m.Text() != "" || m.Media() != nil {
		_, err = tgbot.SendMessage(peer, m.Text(), opts)
	}
	return err
}

func pmBotEnableCmd(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		status := "disabled"
		if isPmBotEnabled() {
			status = "enabled"
		}
		_, err := eOR(m, locales.Trf("pmbot.status", status))
		return err
	}

	switch strings.ToLower(args) {
	case "on", "enable", "true":
		Db.Set(context.Background(), pmBotEnabledKey, "true", 0)
		_, err := eOR(m, locales.Tr("pmbot.enabled"))
		return err
	case "off", "disable", "false":
		Db.Set(context.Background(), pmBotEnabledKey, "false", 0)
		_, err := eOR(m, locales.Tr("pmbot.disabled"))
		return err
	default:
		_, err := eOR(m, locales.Tr("pmbot.enable_usage"))
		return err
	}
}

func pmBotSetLogCmd(m *telegram.NewMessage) error {
	args := m.Args()
	if args == "" {
		_, err := eOR(m, locales.Tr("pmbot.setlog_usage"))
		return err
	}

	chatID := utils.StringToInt64(args)
	if chatID == 0 {
		_, err := eOR(m, locales.Tr("pmbot.invalid_chat"))
		return err
	}

	peer, err := tgbot.GetSendablePeer(chatID)
	if err != nil {
		_, err = eOR(m, locales.Tr("pmbot.bot_not_in_chat"))
		return err
	}

	_, err = tgbot.SendMessage(peer, locales.Tr("pmbot.log_set_confirm"), &telegram.SendOptions{
		ParseMode: "HTML",
	})
	if err != nil {
		_, err = eOR(m, locales.Tr("pmbot.send_error"))
		return err
	}

	Db.Set(context.Background(), pmBotLogGroupKey, chatID, 0)
	_, err = eOR(m, locales.Trf("pmbot.log_set", chatID))
	return err
}

func pmBotBlockCmd(m *telegram.NewMessage) error {
	logGroup := getLogGroup()
	topicID, isTopic := m.TopicID()
	var userID int64

	if isTopic && topicID != 0 && logGroup != 0 && m.ChatID() == logGroup {
		userID = resolveTopicUser(topicID)
	}

	if userID == 0 {
		userID, _ = ExtractUser(m)
	}

	if userID == 0 {
		_, err := eOR(m, locales.Tr("pmbot.block_usage"))
		return err
	}

	blockUser(userID)

	tid := getTopicForUser(userID)
	if tid != 0 && logGroup != 0 {
		peer, err := tgbot.GetSendablePeer(logGroup)
		if err == nil {
			tgbot.MessagesEditForumTopic(&telegram.MessagesEditForumTopicParams{
				Peer:    peer,
				TopicID: tid,
				Closed:  true,
			})
		}
	}

	_, userName := GetUserInfo(userID)
	_, err := eOR(m, locales.Trf("pmbot.user_blocked", userID, userName))
	return err
}

func pmBotUnblockCmd(m *telegram.NewMessage) error {
	logGroup := getLogGroup()
	topicID, isTopic := m.TopicID()
	var userID int64

	if isTopic && topicID != 0 && logGroup != 0 && m.ChatID() == logGroup {
		userID = resolveTopicUser(topicID)
	}

	if userID == 0 {
		userID, _ = ExtractUser(m)
	}

	if userID == 0 {
		_, err := eOR(m, locales.Tr("pmbot.unblock_usage"))
		return err
	}

	unblockUser(userID)

	tid := getTopicForUser(userID)
	if tid != 0 && logGroup != 0 {
		peer, err := tgbot.GetSendablePeer(logGroup)
		if err == nil {
			tgbot.MessagesEditForumTopic(&telegram.MessagesEditForumTopicParams{
				Peer:    peer,
				TopicID: tid,
				Closed:  false,
			})
		}
	}

	_, userName := GetUserInfo(userID)
	_, err := eOR(m, locales.Trf("pmbot.user_unblocked", userID, userName))
	return err
}

func loadPmBotModule() {
	handlers := []*Handler{
		{ModuleName: "PM Bot", Command: "pmbot", Description: locales.Tr("desc.pmbot"), Func: pmBotEnableCmd},
		{ModuleName: "PM Bot", Command: "setpmlog", Description: locales.Tr("desc.setpmlog"), Func: pmBotSetLogCmd},
		{ModuleName: "PM Bot", Command: "pmblock", Description: locales.Tr("desc.pmblock"), Func: pmBotBlockCmd},
		{ModuleName: "PM Bot", Command: "pmunblock", Description: locales.Tr("desc.pmunblock"), Func: pmBotUnblockCmd},
	}
	AddHandlers(handlers, client)

	tgbot.OnCommand("start", onBotStart).Private()
	tgbot.OnMessage(string(telegram.OnNewMessage), onBotPrivateMessage, telegram.IsPrivate)
	tgbot.AddMessageHandler(telegram.OnNewMessage, onLogGroupReply)
}
