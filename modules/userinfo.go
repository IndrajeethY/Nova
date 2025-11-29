package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func parseBirthday(day, month, year int32) string {
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	result := strconv.Itoa(int(day)) + ", " + months[month-1]
	if year != 0 {
		result += ", " + strconv.Itoa(int(year))
	}

	currYear := time.Now().Year()
	bday := time.Date(currYear, time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	if bday.Before(time.Now()) {
		bday = time.Date(currYear+1, time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	}
	days := int(time.Until(bday).Hours() / 24)
	return result + "; is in " + strconv.Itoa(days) + " days"
}

func StatsCmd(m *telegram.NewMessage) error {
	var (
		admingc, adminch, creator, users, bots, grps, channels, deleted, total, notify,
		pinned, blockedc, contacts, mutuals int
		unreadmsgs, mentions, reactions int32
	)

	msg, _ := eOR(m, locales.Tr("userinfo.fetching_stats"))
	client.IterDialogs(func(d *telegram.TLDialog) error {
		dialog, ok := d.Dialog.(*telegram.DialogObj)
		if !ok {
			return nil
		}
		if dialog.NotifySettings != nil && !dialog.NotifySettings.Silent {
			notify++
		}
		unreadmsgs += dialog.UnreadCount
		mentions += dialog.UnreadMentionsCount
		reactions += dialog.UnreadReactionsCount

		if dialog.Pinned {
			pinned++
		}
		switch p := dialog.Peer.(type) {
		case *telegram.PeerChannel:
			total++
			ch, err := client.GetChannel(p.ChannelID)
			if err != nil {
				return nil
			}

			if ch.Creator {
				creator++
			}
			if ch.Broadcast {
				if ch.AdminRights == nil {
					adminch++
				}
				channels++
			} else {
				if ch.AdminRights != nil {
					admingc++
				}
				grps++
			}
		case *telegram.PeerUser:
			total++
			user, err := client.GetUser(p.UserID)
			if err != nil {
				return nil
			}

			if user.Deleted {
				deleted++
			}
			if user.Bot {
				bots++
			} else {
				users++
			}

			if user.MutualContact {
				mutuals++
			}
			if user.Contact {
				contacts++
			}
		case *telegram.PeerChat:
			total++
			chat, err := client.GetChat(p.ChatID)
			if err != nil {
				return nil
			}
			if chat.AdminRights != nil {
				admingc++
			}
			grps++
		}

		return nil
	}, &telegram.DialogOptions{SleepThresholdMs: 10, Limit: 5000})

	blocked, _ := client.ContactsGetBlocked(false, 0, 5000)
	if b, ok := blocked.(*telegram.ContactsBlockedObj); ok {
		blockedc = len(b.Users)
	}
	response := fmt.Sprintf(locales.Tr("userinfo.stats"),
		users, bots, grps, channels, contacts, blockedc,
		pinned, unreadmsgs, notify, mentions, reactions,
		creator, admingc, adminch, deleted, mutuals,
		total, blockedc,
	)

	_, err := msg.Edit(response)
	return err
}

func userInfo(m *telegram.NewMessage) error {
	userId, _, _ := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("userinfo.usage_info"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("userinfo.fetching_info"))
	peer, _ := client.GetSendablePeer(userId)

	response := locales.Tr("userinfo.info_header")
	var photo *telegram.InputMediaPhoto

	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		userinfo, _ := m.Client.UsersGetFullUser(&telegram.InputUserObj{UserID: p.UserID, AccessHash: p.AccessHash})
		uf := userinfo.FullUser
		un := userinfo.Users[0].(*telegram.UserObj)

		if un.FirstName != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.first_name"), un.FirstName)
		}
		if un.LastName != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.last_name"), un.LastName)
		}
		response += fmt.Sprintf(locales.Tr("userinfo.user_id"), un.ID)
		if un.Username != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.username"), un.Username)
		}
		if uf.About != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.about"), uf.About)
		}
		if un.Usernames != nil {
			var names string
			for _, v := range un.Usernames {
				names += "@" + v.Username + " "
			}
			response += fmt.Sprintf(locales.Tr("userinfo.usernames"), names)
		}
		if uf.Birthday != nil {
			response += fmt.Sprintf(locales.Tr("userinfo.birthday"), parseBirthday(uf.Birthday.Day, uf.Birthday.Month, uf.Birthday.Year))
		}
		response += fmt.Sprintf(locales.Tr("userinfo.user_link"), un.ID)

		if uf.ProfilePhoto != nil {
			pic := uf.ProfilePhoto.(*telegram.PhotoObj)
			response += fmt.Sprintf(locales.Tr("userinfo.dc_id"), pic.DcID)
			if uf.PersonalPhoto != nil {
				pic = uf.PersonalPhoto.(*telegram.PhotoObj)
			}
			photo = &telegram.InputMediaPhoto{
				ID:      &telegram.InputPhotoObj{ID: pic.ID, AccessHash: pic.AccessHash, FileReference: pic.FileReference},
				Spoiler: true,
			}
		}

		response += fmt.Sprintf(locales.Tr("userinfo.is_bot"), un.Bot)
		response += fmt.Sprintf(locales.Tr("userinfo.is_deleted"), un.Deleted)
		response += fmt.Sprintf(locales.Tr("userinfo.is_contact"), un.Contact)
		response += fmt.Sprintf(locales.Tr("userinfo.is_mutual"), un.MutualContact)
		response += fmt.Sprintf(locales.Tr("userinfo.is_premium"), un.Premium)

	case *telegram.InputPeerChannel:
		chatInfo, _ := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{ChannelID: p.ChannelID, AccessHash: p.AccessHash})
		cf := chatInfo.FullChat.(*telegram.ChannelFull)
		cobj := chatInfo.Chats[0].(*telegram.Channel)

		response += fmt.Sprintf(locales.Tr("userinfo.chat_title"), cobj.Title)
		response += fmt.Sprintf(locales.Tr("userinfo.chat_id"), cobj.ID)
		if cobj.Username != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.chat_username"), cobj.Username)
		}
		if cf.About != "" {
			response += fmt.Sprintf(locales.Tr("userinfo.about"), cf.About)
		}
		if cf.ChatPhoto != nil {
			pic := cf.ChatPhoto.(*telegram.PhotoObj)
			response += fmt.Sprintf(locales.Tr("userinfo.dc_id"), pic.DcID)
			photo = &telegram.InputMediaPhoto{
				ID:      &telegram.InputPhotoObj{ID: pic.ID, AccessHash: pic.AccessHash, FileReference: pic.FileReference},
				Spoiler: true,
			}
		}
		response += fmt.Sprintf(locales.Tr("userinfo.participants"), cf.ParticipantsCount)
		response += fmt.Sprintf(locales.Tr("userinfo.admins_count"), cf.AdminsCount)

	default:
		response = locales.Tr("userinfo.unknown_peer")
	}

	if photo != nil && photo.ID != nil {
		_, err := m.Client.EditMessage(m.ChatID(), msg.ID, response, &telegram.SendOptions{Media: photo})
		return err
	}
	_, err := msg.Edit(response)
	return err
}

func idCmd(m *telegram.NewMessage) error {
	userId, _, _ := ExtractUserMsg(m)
	if userId == 0 {
		userId = m.SenderID()
	}

	response := fmt.Sprintf(locales.Tr("userinfo.id_user"), userId, userId)
	response += fmt.Sprintf(locales.Tr("userinfo.id_chat"), msgLink(m), m.ChatID())
	response += fmt.Sprintf(locales.Tr("userinfo.id_message"), msgLink(m), m.ID)

	if m.IsReply() {
		reply, _ := m.GetReplyMessage()
		response += fmt.Sprintf(locales.Tr("userinfo.id_reply"), msgLink(reply), reply.ID)
		if reply.File != nil {
			response += fmt.Sprintf(locales.Tr("userinfo.id_file"), msgLink(reply), reply.File.FileID)
		}
	}

	_, err := eOR(m, response)
	return err
}

func LoadMyinfo(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "stats", Description: "Fetch user statistics", Func: StatsCmd, ModuleName: "User Info"},
		{Command: "info", Description: "Fetch user info", Func: userInfo, ModuleName: "User Info"},
		{Command: "id", Description: "Fetch ID info", Func: idCmd, ModuleName: "User Info"},
	}
	AddHandlers(handlers, c)
}
