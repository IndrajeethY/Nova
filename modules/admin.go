package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func BanUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_ban"))
		return err
	}
	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}
	msg, _ := eOR(m, locales.Tr("admin.banning"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Ban: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.ban_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.banned"), userId, userName, reason))
	return err
}

func UnbanUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_unban"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("admin.unbanning"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Unban: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.unban_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.unbanned"), userId, userName))
	return err
}

func KickUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_kick"))
		return err
	}
	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}
	msg, _ := eOR(m, locales.Tr("admin.kicking"))
	_, err := m.Client.KickParticipant(m.ChatID(), userId)
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.kick_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.kicked"), userId, userName, reason))
	return err
}

func MuteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_mute"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("admin.muting"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Mute: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.mute_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.muted"), userId, userName))
	return err
}

func UnmuteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_unmute"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("admin.unmuting"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Unmute: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.unmute_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.unmuted"), userId, userName))
	return err
}

func DmuteUser(m *telegram.NewMessage) error {
	userId, userName, _ := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_dmute"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("admin.dmuting"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Mute: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.dmute_error"))
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.muted"), userId, userName))
	return err
}

func DkickUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_kick"))
		return err
	}
	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}
	msg, _ := eOR(m, locales.Tr("admin.kicking"))
	_, err := m.Client.KickParticipant(m.ChatID(), userId)
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.kick_error"))
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.kicked"), userId, userName, reason))
	return err
}

func DbanUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_ban"))
		return err
	}
	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}
	msg, _ := eOR(m, locales.Tr("admin.banning"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Ban: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.ban_error"))
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.banned"), userId, userName, reason))
	return err
}

func PromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_promote"))
		return err
	}
	if title == "" {
		title = "Λ∂мιи"
	}
	msg, _ := eOR(m, locales.Tr("admin.promoting"))
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{Rights: &telegram.ChatAdminRights{ChangeInfo: true, DeleteMessages: true, InviteUsers: true, BanUsers: true, PinMessages: true}, Rank: title, IsAdmin: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.promote_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.promoted"), userId, userName))
	return err
}

func FullPromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_promote"))
		return err
	}
	if title == "" {
		title = "𝙎υρєя Λ∂мιи"
	}
	msg, _ := eOR(m, locales.Tr("admin.promoting"))
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{Rights: &telegram.ChatAdminRights{ChangeInfo: true, DeleteMessages: true, InviteUsers: true, BanUsers: true, PinMessages: true, AddAdmins: true, ManageCall: true, ManageTopics: true, PostStories: true, EditStories: true, DeleteStories: true}, Rank: title, IsAdmin: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.promote_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.promoted"), userId, userName))
	return err
}

func DemoteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_demote"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("admin.demoting"))
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{IsAdmin: false, Rights: &telegram.ChatAdminRights{}})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.demote_error"))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.demoted"), userId, userName))
	return err
}

func PinMessage(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("admin.groups_only"))
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_pin"))
		return err
	}
	var silent bool
	if strings.Contains(m.Args(), "silent") {
		silent = true
	}
	err = reply.Pin(&telegram.PinOptions{Silent: silent})
	if err != nil {
		_, err := eOR(m, locales.Tr("admin.pin_error"))
		return err
	}
	_, err = eOR(m, fmt.Sprintf(locales.Tr("admin.pinned"), msgLink(reply)))
	return err
}

func UnpinMessage(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("admin.groups_only"))
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_unpin"))
		return err
	}
	err = reply.Pin(&telegram.PinOptions{Unpin: true})
	if err != nil {
		_, err := eOR(m, locales.Tr("admin.unpin_error"))
		return err
	}
	_, err = eOR(m, fmt.Sprintf(locales.Tr("admin.unpinned"), msgLink(reply)))
	return err
}

func zombiesCmd(m *telegram.NewMessage) error {
	args := m.Args()
	if !m.IsGroup() {
		_, err := m.Edit(locales.Tr("admin.groups_only"))
		return err
	}
	perms, err := m.Client.GetChatMember(m.ChatID(), m.Sender.ID)
	if err != nil {
		_, err = m.Edit(locales.Tr("admin.get_permissions_error"))
		return err
	}
	if !perms.Rights.BanUsers {
		_, err = m.Edit(locales.Tr("admin.no_permissions"))
		return err
	}
	deleted := []int64{}
	msg, _ := m.Edit(locales.Tr("admin.zombies_searching"))
	members, _, err := m.Client.GetChatMembers(m.ChatID(), &telegram.ParticipantOptions{Limit: 500000})
	if err != nil {
		_, err = msg.Edit(locales.Tr("admin.get_permissions_error"))
		return err
	}
	for _, member := range members {
		if member.User.Bot {
			continue
		}
		if member.User.Deleted {
			deleted = append(deleted, member.User.ID)
		}
	}
	if len(deleted) == 0 {
		_, err = msg.Edit(locales.Tr("admin.zombies_not_found"))
		return err
	}
	if strings.Contains(args, "clean") && len(deleted) > 0 {
		success := 0
		failed := 0
		for _, id := range deleted {
			_, err := m.Client.KickParticipant(m.ChatID(), id)
			if err != nil {
				failed++
			} else {
				success++
			}
		}
		_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.zombies_cleaned"), success, failed))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("admin.zombies_found"), len(deleted)))
	return err
}

func LoadAdminModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "ban", Func: BanUser, Description: "Ban a user from the chat", ModuleName: "Admin"},
		{Command: "unban", Func: UnbanUser, Description: "Unban a user from the chat", ModuleName: "Admin"},
		{Command: "kick", Func: KickUser, Description: "Kick a user from the chat", ModuleName: "Admin"},
		{Command: "mute", Func: MuteUser, Description: "Mute a user in the chat", ModuleName: "Admin"},
		{Command: "unmute", Func: UnmuteUser, Description: "Unmute a user in the chat", ModuleName: "Admin"},
		{Command: "dmute", Func: DmuteUser, Description: "Demote and mute a user", ModuleName: "Admin"},
		{Command: "demote", Func: DemoteUser, Description: "Demote a user from admin", ModuleName: "Admin"},
		{Command: "dkick", Func: DkickUser, Description: "Demote and kick a user", ModuleName: "Admin"},
		{Command: "dban", Func: DbanUser, Description: "Demote and ban a user", ModuleName: "Admin"},
		{Command: "promote", Func: PromoteUser, Description: "Promote a user to admin", ModuleName: "Admin"},
		{Command: "fullpromote", Func: FullPromoteUser, Description: "Fully promote a user", ModuleName: "Admin"},
		{Command: "pin", Func: PinMessage, Description: "Pin a message in the chat", ModuleName: "Admin"},
		{Command: "unpin", Func: UnpinMessage, Description: "Unpin a message", ModuleName: "Admin"},
		{Command: "zombies", Func: zombiesCmd, Description: "Find and clean deleted accounts", ModuleName: "Admin"},
	}
	AddHandlers(handlers, c)
}
