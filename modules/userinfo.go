package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"strconv"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("User Info", loadUserInfoModule)
}

func parseBirthday(dat, month, year int32) string {
	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
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
	if timeBday.Before(time.Now()) {
		timeBday = time.Date(currYear+1, time.Month(month), int(dat), 0, 0, 0, 0, time.UTC)
	}
	days := timeBday.Sub(time.Now()).Hours() / 24
	return strconv.Itoa(int(days)) + " days"
}

func StatsCmd(m *telegram.NewMessage) error {
	var (
		admingc, adminch, creator, users, bots, grps, channels, deleted, total, notify, pinned, blockedc, contacts, mutuals int
		unreadmsgs, mentions, reactions                                                                                     int32
	)
	msg, _ := eOR(m, locales.Tr("userinfo.fetching_stats"))
	client.IterDialogs(func(d *telegram.TLDialog) error {
		if dlg, ok := d.Dialog.(*telegram.DialogObj); ok {
			if !dlg.NotifySettings.Silent {
				notify++
			}
			if dlg.UnreadCount > 0 {
				unreadmsgs += dlg.UnreadCount
			}
			if dlg.UnreadMentionsCount > 0 {
				mentions += dlg.UnreadMentionsCount
			}
			if dlg.UnreadReactionsCount > 0 {
				reactions += dlg.UnreadReactionsCount
			}
			if dlg.Pinned {
				pinned++
			}
		}
		switch p := d.Peer.(type) {
		case *telegram.PeerChannel:
			total++
			chatInfo, err := client.GetChannel(p.ChannelID)
			if err != nil {
				return nil
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
				return nil
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
				return nil
			}
			if chatInfo.AdminRights != nil {
				admingc++
			}
			grps++
		}
		return nil
	}, &telegram.DialogOptions{SleepThresholdMs: 10, Limit: 5000})
	blocked, err := client.ContactsGetBlocked(false, 0, 5000)
	if err == nil {
		if obj, ok := blocked.(*telegram.ContactsBlockedObj); ok {
			blockedc = len(obj.Users)
		}
	}
	response := locales.Trf("userinfo.stats",
		users, bots, grps, channels, contacts, blockedc, pinned, unreadmsgs, notify, mentions, reactions,
		creator, admingc, adminch, deleted, mutuals, total, blockedc,
	)
	_, err = msg.Edit(response)
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
		userinfo, _ := m.Client.UsersGetFullUser(&telegram.InputUserObj{
			UserID: p.UserID, AccessHash: p.AccessHash,
		})
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
			var s string
			for _, v := range un.Usernames {
				s += "@" + v.Username + " "
			}
			response += fmt.Sprintf(locales.Tr("userinfo.usernames"), s)
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
				ID: &telegram.InputPhotoObj{
					ID: pic.ID, AccessHash: pic.AccessHash, FileReference: pic.FileReference,
				},
				Spoiler: true,
			}
		}
		response += fmt.Sprintf(locales.Tr("userinfo.is_bot"), un.Bot)
		response += fmt.Sprintf(locales.Tr("userinfo.is_deleted"), un.Deleted)
		response += fmt.Sprintf(locales.Tr("userinfo.is_contact"), un.Contact)
		response += fmt.Sprintf(locales.Tr("userinfo.is_mutual"), un.MutualContact)
		response += fmt.Sprintf(locales.Tr("userinfo.is_premium"), un.Premium)

	case *telegram.InputPeerChannel:
		chatInfo, _ := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{
			ChannelID: p.ChannelID, AccessHash: p.AccessHash,
		})
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
				ID: &telegram.InputPhotoObj{
					ID: pic.ID, AccessHash: pic.AccessHash, FileReference: pic.FileReference,
				},
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
		if reply != nil {
			response += fmt.Sprintf(locales.Tr("userinfo.id_reply"), msgLink(reply), reply.ID)
			if reply.File != nil {
				response += fmt.Sprintf(locales.Tr("userinfo.id_file"), msgLink(reply), reply.File.FileID)
			}
		}
	}
	_, err := eOR(m, response)
	return err
}

func deletePfpCmd(m *telegram.NewMessage) error {
	userID, _, _ := ExtractUserMsg(m)
	if userID == 0 {
		userID = m.SenderID()
	}
	pics, _ := client.GetProfilePhotos(userID, &telegram.PhotosOptions{Limit: 10000})
	if len(pics) == 0 {
		_, _ = eOR(m, locales.Tr("userinfo.no_photos"))
		return nil
	}
	allPics := make([]telegram.InputPhoto, 0, len(pics))
	for _, p := range pics {
		data := p.Photo.(*telegram.PhotoObj)
		allPics = append(allPics, &telegram.InputPhotoObj{
			ID: data.ID, AccessHash: data.AccessHash, FileReference: data.FileReference,
		})
	}
	_, err := m.Client.PhotosDeletePhotos(allPics)
	if err != nil {
		_, _ = eOR(m, locales.Trf("userinfo.photos_error", err))
		return err
	}
	_, _ = eOR(m, locales.Tr("userinfo.photos_deleted"))
	return nil
}

func myGroupsCmd(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("userinfo.groups_fetching"))
	var result string
	count := 0
	client.IterDialogs(func(d *telegram.TLDialog) error {
		switch p := d.Peer.(type) {
		case *telegram.PeerChannel:
			chatInfo, err := client.GetChannel(p.ChannelID)
			if err != nil || !chatInfo.Creator || chatInfo.Broadcast {
				return nil
			}
			result += fmt.Sprintf(locales.Tr("userinfo.groups_entry"), chatInfo.Title, chatInfo.ID)
			count++
		case *telegram.PeerChat:
			chatInfo, err := client.GetChat(p.ChatID)
			if err != nil || !chatInfo.Creator {
				return nil
			}
			result += fmt.Sprintf(locales.Tr("userinfo.groups_entry"), chatInfo.Title, chatInfo.ID)
			count++
		}
		return nil
	}, &telegram.DialogOptions{SleepThresholdMs: 10, Limit: 5000})
	if count == 0 {
		result = locales.Tr("userinfo.groups_none")
	}
	_, err := msg.Edit(locales.Tr("userinfo.groups_header") + result)
	return err
}

func myChannelsCmd(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("userinfo.channels_fetching"))
	var result string
	count := 0
	client.IterDialogs(func(d *telegram.TLDialog) error {
		if p, ok := d.Peer.(*telegram.PeerChannel); ok {
			chatInfo, err := client.GetChannel(p.ChannelID)
			if err != nil || !chatInfo.Creator || !chatInfo.Broadcast {
				return nil
			}
			result += fmt.Sprintf(locales.Tr("userinfo.channels_entry"), chatInfo.Title, chatInfo.ID)
			if chatInfo.Username != "" {
				result += fmt.Sprintf("t.me/%s\n", chatInfo.Username)
			}
			count++
		}
		return nil
	}, &telegram.DialogOptions{SleepThresholdMs: 10, Limit: 5000})
	if count == 0 {
		result = locales.Tr("userinfo.channels_none")
	}
	_, err := msg.Edit(locales.Tr("userinfo.channels_header") + result)
	return err
}

func leftChatsCmd(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("userinfo.leftchats_fetching"))
	leftChatsObj, err := client.ChannelsGetLeftChannels(10000)
	if err != nil || leftChatsObj == nil {
		_, err2 := msg.Edit(locales.Tr("userinfo.leftchats_none"))
		return err2
	}
	var result string
	leftChats := leftChatsObj.(*telegram.MessagesChatsObj)
	for _, chatObj := range leftChats.Chats {
		switch chat := chatObj.(type) {
		case *telegram.ChatObj:
			if chat.Title != "" {
				result += fmt.Sprintf("<b>%s</b> | <code>%d</code>\n", chat.Title, chat.ID)
			}
		case *telegram.Channel:
			if chat.Title != "" {
				result += fmt.Sprintf("<b>%s</b> | <code>%d</code>\n", chat.Title, chat.ID)
			}
		}
	}
	if result == "" {
		_, err2 := msg.Edit(locales.Tr("userinfo.leftchats_none"))
		return err2
	}
	_, err2 := msg.Edit(locales.Tr("userinfo.leftchats_header") + result)
	return err2
}

func loadUserInfoModule() {
	handlers := []*Handler{
		{Command: "stats", Description: "Fetch complete stats", Func: StatsCmd, ModuleName: "User Info"},
		{Command: "info", Description: "Fetch info about a user", Func: userInfo, ModuleName: "User Info"},
		{Command: "id", Description: "Fetch ID of user or sender", Func: idCmd, ModuleName: "User Info"},
		{Command: "deletepfp", Description: "Delete all profile photos", Func: deletePfpCmd, ModuleName: "User Info"},
		{Command: "mygroups", Description: "Fetch owned groups", Func: myGroupsCmd, ModuleName: "User Info"},
		{Command: "mychannels", Description: "Fetch owned channels", Func: myChannelsCmd, ModuleName: "User Info"},
		{Command: "leftchats", Description: "Fetch left chats", Func: leftChatsCmd, ModuleName: "User Info"},
	}
	AddHandlers(handlers, client)
}
