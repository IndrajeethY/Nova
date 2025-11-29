package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"fmt"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

// ============================================================================
// Logging Helpers
// ============================================================================

type TgLogSender struct{}

func (t *TgLogSender) SendLog(msg string) error {
	return logMessage(msg)
}

func logMessage(msg string) error {
	logChatStr := db.Get("LOG_CHAT")
	logChat := utils.StringToInt64(logChatStr)
	if logChat == 0 {
		logChat = ubId
	}

	peer, err := tgbot.GetSendablePeer(logChat)
	if err != nil {
		logger.Errorf("logMessage peer error: %v", err)
		return err
	}

	_, err = tgbot.SendMessage(peer, msg)
	return err
}

// ============================================================================
// Permission Helpers
// ============================================================================

func IsAdmin(user, chat int64) bool {
	perms, err := client.GetChatMember(chat, user)
	if err != nil {
		return false
	}
	return perms.Status == "creator" || perms.Status == "administrator"
}

// ============================================================================
// Message Helpers
// ============================================================================

func msgLink(m *telegram.NewMessage) string {
	if m.IsPrivate() {
		return ""
	}
	if m.Channel.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", m.Channel.Username, m.ID)
	}
	return fmt.Sprintf("https://t.me/c/%d/%d", m.ChatID(), m.ID)
}

// eOR edits message if sent by owner, otherwise replies
func eOR(m *telegram.NewMessage, text string, opts ...telegram.SendOptions) (*telegram.NewMessage, error) {
	ptrs := make([]*telegram.SendOptions, len(opts))
	for i := range opts {
		ptrs[i] = &opts[i]
	}
	if m.Sender.ID == ubId {
		return m.Edit(text, ptrs...)
	}
	return m.Reply(text, ptrs...)
}

// ============================================================================
// User Extraction Helpers
// ============================================================================

// ExtractUserMsg extracts user ID, name and remaining message from command arguments
func ExtractUserMsg(m *telegram.NewMessage) (int64, string, string) {
	var userId int64
	var userName, msg string
	msg = m.Args()

	if m.IsReply() {
		replied, err := m.GetReplyMessage()
		if err != nil {
			return 0, "", ""
		}
		return replied.Sender.ID, replied.Sender.FirstName + " " + replied.Sender.LastName, msg
	}

	re := regexp.MustCompile(`^(\d+)|@(\w+)|https://t.me/(\w+)|tg://user\?id=(\d+)`)
	matches := re.FindStringSubmatch(msg)

	if len(matches) > 0 {
		splited := strings.Split(msg, " ")
		switch {
		case matches[1] != "":
			userId, userName = GetUserInfo(utils.StringToInt64(matches[1]))
			msg = strings.Replace(msg, matches[1], "", 1)
		case matches[2] != "":
			userId, userName = GetUserInfo(matches[2])
			msg = strings.Replace(msg, splited[0], "", 1)
		case matches[3] != "":
			userId, userName = GetUserInfo(matches[3])
			msg = strings.Replace(msg, splited[0], "", 1)
		case matches[4] != "":
			userId, userName = GetUserInfo(utils.StringToInt64(matches[4]))
			msg = strings.Replace(msg, splited[0], "", 1)
		}
	} else if len(m.Message.Entities) > 0 {
		for _, entity := range m.Message.Entities {
			if ent, ok := entity.(*telegram.MessageEntityTextURL); ok {
				if strings.Contains(ent.URL, "t.me/") {
					parts := strings.Split(ent.URL, "/")
					userId, userName = GetUserInfo(parts[len(parts)-1])
					msg = strings.Replace(msg, ent.URL, "", 1)
				} else if strings.Contains(ent.URL, "tg://user?id=") {
					idStr := strings.TrimPrefix(ent.URL, "tg://user?id=")
					userId, userName = GetUserInfo(utils.StringToInt64(idStr))
					msg = strings.Replace(msg, ent.URL, "", 1)
				}
				break
			}
		}
	}

	return userId, userName, strings.TrimSpace(msg)
}

// ExtractUser extracts only user ID and name from command arguments
func ExtractUser(m *telegram.NewMessage) (int64, string) {
	userId, userName, _ := ExtractUserMsg(m)
	return userId, userName
}

// GetUserInfo retrieves user/chat info by ID or username
func GetUserInfo(userId any) (int64, string) {
	peer, err := client.GetSendablePeer(userId)
	if err != nil {
		return 0, ""
	}

	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		user, err := client.GetUser(p.UserID)
		if err != nil {
			return 0, ""
		}
		return user.ID, strings.TrimSpace(user.FirstName + " " + user.LastName)
	case *telegram.InputPeerChannel:
		ch, err := client.GetChannel(p.ChannelID)
		if err != nil {
			return 0, ""
		}
		return ch.ID, ch.Title
	case *telegram.InputPeerChat:
		chat, err := client.GetChat(p.ChatID)
		if err != nil {
			return 0, ""
		}
		return chat.ID, chat.Title
	}
	return 0, ""
}
