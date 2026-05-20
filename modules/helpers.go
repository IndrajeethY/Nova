package modules

import (
	"NovaUserbot/utils"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	log "github.com/sirupsen/logrus"
)

var extractUserRe = regexp.MustCompile(`^(?:(\d+)|@(\w+)|https://t.me/(\w+)|tg://user\?id=(\d+))`)

func IsAdmin(user int64, chat int64) bool {
	perms, err := client.GetChatMember(chat, user)
	if err != nil {
		log.Println("Error getting chat member:", err)
		return false
	}
	return perms.Status == "creator" || perms.Status == "administrator"
}

func msgLink(m *telegram.NewMessage) string {
	if m.IsPrivate() {
		return ""
	}
	if m.Channel.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", m.Channel.Username, m.ID)
	}
	return fmt.Sprintf("https://t.me/c/%d/%d", m.ChatID(), m.ID)
}

func eOR(m *telegram.NewMessage, text string, opts ...*telegram.SendOptions) (*telegram.NewMessage, error) {
	if m.Sender.ID == ubId {
		return m.Edit(text, opts...)
	}
	return m.Reply(text, opts...)
}

func ExtractUserMsg(m *telegram.NewMessage) (int64, string, string) {
	var userId int64
	var userName, msg string
	msg = m.Args()
	if m.IsReply() {
		replied, err := m.GetReplyMessage()
		if err != nil {
			return 0, "", ""
		}
		userId = replied.Sender.ID
		userName = replied.Sender.FirstName + " " + replied.Sender.LastName
	} else {
		matches := extractUserRe.FindStringSubmatch(msg)
		if len(matches) > 0 {
			splited := strings.Split(msg, " ")
			if matches[1] != "" {
				userId, userName = GetUserInfo(utils.StringToInt64(matches[1]))
				msg = strings.Replace(msg, matches[1], "", 1)
			} else if matches[2] != "" {
				userId, userName = GetUserInfo(matches[2])
				msg = strings.Replace(msg, splited[0], "", 1)
			} else if matches[3] != "" {
				userId, userName = GetUserInfo(matches[3])
				msg = strings.Replace(msg, splited[0], "", 1)
			} else if matches[4] != "" {
				userId, userName = GetUserInfo(utils.StringToInt64(matches[4]))
				msg = strings.Replace(msg, splited[0], "", 1)
			}
		} else if len(m.Message.Entities) > 0 {
			for _, entity := range m.Message.Entities {
				if ent, ok := entity.(*telegram.MessageEntityTextURL); ok {
					if strings.Contains(ent.URL, "t.me/") || strings.Contains(ent.URL, "https://t.me/") {
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
		if msg == "" && userName == "" && userId == 0 {
			return 0, "", ""
		}
	}

	return userId, userName, strings.TrimSpace(msg)
}

func ExtractUser(m *telegram.NewMessage) (int64, string) {
	userId, userName, _ := ExtractUserMsg(m)
	return userId, userName
}

func GetUserInfo(userId any) (int64, string) {
	peer, err := client.GetSendablePeer(userId)
	if err != nil {
		log.Println("Error getting sendable peer:", err)
		return 0, ""
	}
	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		userinfo, err := client.GetUser(p.UserID)
		if err != nil {
			return 0, ""
		}
		return userinfo.ID, userinfo.FirstName + " " + userinfo.LastName
	case *telegram.InputPeerChannel:
		chatinfo, err := client.GetChannel(p.ChannelID)
		if err != nil {
			return 0, ""
		}
		return chatinfo.ID, chatinfo.Title
	case *telegram.InputPeerChat:
		chatinfo, err := client.GetChat(p.ChatID)
		if err != nil {
			return 0, ""
		}
		return chatinfo.ID, chatinfo.Title
	default:
		return 0, ""
	}
}

func logMessage(m string) error {
	var logChat int64
	logChatConfig, err := Db.Get(context.Background(), "LOG_CHAT").Result()
	if err != nil {
		logChat = ubId
	} else {
		logChat = utils.StringToInt64(logChatConfig)
	}
	peer, err := tgbot.GetSendablePeer(logChat)
	if err != nil {
		log.Println("Error getting sendable peer:", err)
		return err
	}
	_, err = tgbot.SendMessage(peer, m)
	if err != nil {
		log.Println("Error sending message to chat:", err)
		return err
	}
	return nil
}
