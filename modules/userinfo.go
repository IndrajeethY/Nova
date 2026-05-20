package modules

import (
	"fmt"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func parseBirthday(dat, month, year int32) string {
	months := []string{
		"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December",
	}
	result := strconv.Itoa(int(dat)) + ", " + months[month-1]
	if year != 0 {
		result += ", " + strconv.Itoa(int(year))
	}

	return result + "; is in " + tillDate(dat, month)
}

func tillDate(dat, month int32) string {
	currYear := time.Now().Year()

	timeBday := time.Date(currYear, time.Month(month), int(dat), 0, 0, 0, 0, time.UTC)
	currTime := time.Now()

	if timeBday.Before(currTime) {
		timeBday = time.Date(currYear+1, time.Month(month), int(dat), 0, 0, 0, 0, time.UTC)
	}

	days := timeBday.Sub(currTime).Hours() / 24

	return strconv.Itoa(int(days)) + " days"
}

func StatsCmd(m *telegram.NewMessage) error {
	var (
		admingc, adminch, creator, users, bots, grps, channels, deleted, total, notify, pinned, blockedc, contacts, mutuals int
		unreadmsgs, mentions, reactions                                                                                     int32
	)
	msg, _ := eOR(m, "<code>Fetching Stats...This may take a while</code>")
	dialogs, _ := client.IterDialogs(&telegram.DialogOptions{SleepThresholdMs: 10, Limit: 5000})
	for v := range dialogs {
		if dialogs, ok := v.(*telegram.DialogObj); ok {
			if !dialogs.NotifySettings.Silent {
				notify++
			}
			if dialogs.UnreadCount > 0 {
				unreadmsgs += dialogs.UnreadCount
			}
			if dialogs.UnreadMentionsCount > 0 {
				mentions += dialogs.UnreadMentionsCount
			}
			if dialogs.UnreadReactionsCount > 0 {
				reactions += dialogs.UnreadReactionsCount
			}
			if dialogs.Pinned {
				pinned++
			}
			switch p := dialogs.Peer.(type) {
			case *telegram.PeerChannel:
				total++
				chatInfo, err := client.GetChannel(p.ChannelID)
				if err != nil {
					continue
				}
				if chatInfo.Creator {
					creator++
				}
				if chatInfo.Broadcast {
					if chatInfo.AdminRights == nil {
						adminch++
					}
					channels++
				} else {
					if chatInfo.AdminRights != nil {
						admingc++
					}
					grps++
				}
			case *telegram.PeerUser:
				total++
				userInfo, err := client.GetUser(p.UserID)
				if err != nil {
					continue
				}
				if userInfo.Deleted {
					deleted++
				}
				if userInfo.Bot {
					bots++
				} else {
					users++
				}
				if userInfo.MutualContact {
					mutuals++
				}
				if userInfo.Contact {
					contacts++
				}

			case *telegram.PeerChat:
				total++
				chatInfo, err := client.GetChat(p.ChatID)
				if err != nil {
					continue
				}
				if chatInfo.AdminRights != nil {
					admingc++
				}
				grps++
			default:
				continue
			}

		}
	}
	blocked, err := client.ContactsGetBlocked(false, 0, 5000)
	if err != nil {
		blockedc = 0
	}
	blockedc = len(blocked.(*telegram.ContactsBlockedObj).Users)
	response := fmt.Sprintf("<b> ‹Usᴇʀ Sᴛᴀᴛs›</b>\n\n<b>👤 Usᴇʀs:</b> %d\n<b>🤖 Bᴏᴛs:</b> %d\n<b>👥 Gʀᴏᴜᴘs:</b> %d\n<b>📡 Cʜᴀɴɴᴇʟs:</b> %d\n<b>📞 Cᴏɴᴛᴀᴄᴛs:</b> %d\n<b>🔒 Bʟᴏᴄᴋᴇᴅ:</b> %d\n<b>📌 Pɪɴɴᴇᴅ:</b> %d\n<b>📩 Uɴʀᴇᴀᴅ Mᴇssᴀɢᴇs:</b> %d\n<b>🔔 Nᴏᴛɪғɪᴄᴀᴛɪᴏɴs:</b> %d\n<b>📢 Mᴇɴᴛɪᴏɴs:</b> %d\n<b>👍 Rᴇᴀᴄᴛɪᴏɴs:</b> %d\n\n<b> ‹Cʜᴀᴛ Sᴛᴀᴛs›</b>\n\n<b>👑 Cʀᴇᴀᴛᴏʀs:</b> %d\n<b>🛡️ Aᴅᴍɪɴs ɪɴ Gʀᴏᴜᴘs:</b> %d\n<b>🛡️ Aᴅᴍɪɴs ɪɴ Cʜᴀɴɴᴇʟs:</b> %d\n<b>🔇 Dᴇʟᴇᴛᴇᴅ Usᴇʀs:</b> %d\n<b>🤝 Mᴜᴛᴜᴀʟ Cᴏɴᴛᴀᴄᴛs:</b> %d\n\n<b> ‹Oᴛʜᴇʀ Sᴛᴀᴛs›</b>\n\n<b>📈 Tᴏᴛᴀʟ Dɪᴀʟᴏɢs:</b> %d\n<b>🚫 Bʟᴏᴄᴋᴇᴅ Cᴏɴᴛᴀᴄᴛs:</b> %d", users, bots, grps, channels, contacts, blockedc, pinned, unreadmsgs, notify, mentions, reactions, creator, admingc, adminch, deleted, mutuals, total, blockedc)
	_, err = msg.Edit(response)
	return err
}

func userInfo(m *telegram.NewMessage) error {
	userId, _, _ := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .info &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Fetching user info...</code>")
	peer, _ := client.GetSendablePeer(userId)
	response := "<b>Fᴇᴛᴄʜᴇᴅ Iɴғᴏ:\n</b>"
	var photo *telegram.InputMediaPhoto
	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		userinfo, _ := m.Client.UsersGetFullUser(&telegram.InputUserObj{
			UserID:     p.UserID,
			AccessHash: p.AccessHash,
		})
		uf := userinfo.FullUser
		un := userinfo.Users[0].(*telegram.UserObj)
		if un.FirstName != "" {
			response += fmt.Sprintf("<b>Fɪʀsᴛ Nᴀᴍᴇ:</b> %s\n", un.FirstName)
		}
		if un.LastName != "" {
			response += fmt.Sprintf("<b>Lᴀsᴛ Nᴀᴍᴇ:</b> %s\n", un.LastName)
		}
		response += fmt.Sprintf("<b>Usᴇʀ Iᴅ:</b> <code>%d</code>\n", un.ID)
		if un.Username != "" {
			response += "<b>Usᴇʀɴᴀᴍᴇ:</b> @" + un.Username + "\n"
		}
		if uf.About != "" {
			response += "<b>Aʙᴏᴜᴛ:</b> <code>" + uf.About + "</code>\n"
		}
		if un.Usernames != nil {
			response += "<b>Usᴇʀɴᴀᴍᴇs:</b> [<code>" + func() string {
				var s string
				for _, v := range un.Usernames {
					s += "@" + v.Username + " "
				}
				return s
			}() + "</code>]\n"
		}

		if uf.Birthday != nil {
			response += fmt.Sprintf("\n<b>Bɪʀᴛʜᴅᴀʏ:</b> %s\n", parseBirthday(uf.Birthday.Day, uf.Birthday.Month, uf.Birthday.Year))
		}

		response += fmt.Sprintf("<b>Usᴇʀ Lɪɴᴋ:</b> <a href='tg://user?id=%d'>Lɪɴᴋ</a>\n", un.ID)
		if uf.ProfilePhoto != nil {
			pic := uf.ProfilePhoto.(*telegram.PhotoObj)
			response += fmt.Sprintf("<b>Dᴄ Iᴅ:</b> <code>%d</code>\n", pic.DcID)
			if uf.PersonalPhoto != nil {
				pic = uf.PersonalPhoto.(*telegram.PhotoObj)
			}
			photo = &telegram.InputMediaPhoto{
				ID: &telegram.InputPhotoObj{
					ID:            pic.ID,
					AccessHash:    pic.AccessHash,
					FileReference: pic.FileReference,
				},
				Spoiler: true,
			}
		}
		response += fmt.Sprintf("<b>Is Bᴏᴛ:</b> <code>%t</code>\n", un.Bot)
		response += fmt.Sprintf("<b>Is Dᴇʟᴇᴛᴇᴅ:</b> <code>%t</code>\n", un.Deleted)
		response += fmt.Sprintf("<b>Is Cᴏɴᴛᴀᴄᴛ:</b> <code>%t</code>\n", un.Contact)
		response += fmt.Sprintf("<b>Is Mᴜᴛᴜᴀʟ Cᴏɴᴛᴀᴄᴛ:</b> <code>%t</code>\n", un.MutualContact)
		response += fmt.Sprintf("<b>Is Pʀᴇᴍɪᴜᴍ:</b> <code>%t</code>\n", un.Premium)

	case *telegram.InputPeerChannel:
		chatInfo, _ := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		})
		cf := chatInfo.FullChat.(*telegram.ChannelFull)
		cobj := chatInfo.Chats[0].(*telegram.Channel)
		response += fmt.Sprintf("<b>Cʜᴀᴛ Tɪᴛʟᴇ:</b> %s\n", cobj.Title)
		response += fmt.Sprintf("<b>Cʜᴀᴛ Iᴅ:</b> <code>%d</code>\n", cobj.ID)
		if cobj.Username != "" {
			response += fmt.Sprintf("<b>Cʜᴀᴛ Usᴇʀɴᴀᴍᴇ:</b> @%s\n", cobj.Username)
		}
		if cf.About != "" {
			response += fmt.Sprintf("<b>Aʙᴏᴜᴛ:</b> <code>%s</code>\n", cf.About)
		}
		if cf.ChatPhoto != nil {
			pic := cf.ChatPhoto.(*telegram.PhotoObj)
			response += fmt.Sprintf("<b>Dᴄ Iᴅ:</b> <code>%d</code>\n", pic.DcID)
			photo = &telegram.InputMediaPhoto{
				ID: &telegram.InputPhotoObj{
					ID:            pic.ID,
					AccessHash:    pic.AccessHash,
					FileReference: pic.FileReference,
				},
				Spoiler: true,
			}
		}
		response += fmt.Sprintf("<b>Pᴀʀᴛɪᴄɪᴘᴀɴᴛs Cᴏᴜɴᴛ:</b> %d\n", cf.ParticipantsCount)
		response += fmt.Sprintf("<b>Aᴅᴍɪɴs Cᴏᴜɴᴛ:</b> %d\n", cf.AdminsCount)
	default:
		response = "<code>Unknown Peer Type</code>"

	}
	if photo.ID != nil {
		_, err := m.Client.EditMessage(m.ChatID(), msg.ID, response, &telegram.SendOptions{Media: photo})
		return err
	} else {
		_, err := msg.Edit(response)
		return err
	}
}

func idCmd(m *telegram.NewMessage) error {
	userId, _, _ := ExtractUserMsg(m)
	if userId == 0 {
		userId = m.SenderID()
	}
	response := fmt.Sprintf("<b><a href='tg://user?id=%d'>User ID</a></b>: <code>%d</code>\n", userId, userId)
	response += fmt.Sprintf("<b><a href='%s'>Chat ID</a></b>: <code>%d</code>\n", msgLink(m), m.ChatID())
	response += fmt.Sprintf("<b><a href='%s'>Message ID</a></b>: <code>%d</code>\n", msgLink(m), m.ID)
	if m.IsReply() {
		reply, _ := m.GetReplyMessage()
		response += fmt.Sprintf("<b><a href='%s'>Replied Message ID</a></b>: <code>%d</code>\n", msgLink(reply), reply.ID)
		if reply.File != nil {
			response += fmt.Sprintf("<b><a href='%s'>Replied File ID</a></b>: <code>%s</code>\n", msgLink(reply), reply.File.FileID)
		}
	}
	_, err := eOR(m, response)
	return err
}

func LoadMyinfo(c *telegram.Client) {
	handlers := []*Handler{
		{
			Command:     "stats",
			Description: "Fetch complete stats about user",
			Func:        StatsCmd,
			ModuleName:  "User Info",
		},
		{
			Command:     "info",
			Description: "Fetch info about a user",
			Func:        userInfo,
			ModuleName:  "User Info",
		},
		{
			Command:     "id",
			Description: "Fetch ID of user or sender",
			Func:        idCmd,
			ModuleName:  "User Info",
		},
	}
	AddHandlers(handlers, c)
}
