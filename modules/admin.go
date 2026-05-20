package modules

import (
	"NovaUserbot/locales"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

var purgeFromMap = struct {
	sync.RWMutex
	m map[int64]int32
}{m: make(map[int64]int32)}

func init() {
	RegisterModule("Admin", loadAdminModule)
}

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
	_, err = msg.Edit(locales.Trf("admin.banned", userId, userName, reason))
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
	_, err = msg.Edit(locales.Trf("admin.unbanned", userId, userName))
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
	_, err = msg.Edit(locales.Trf("admin.kicked", userId, userName, reason))
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
	_, err = msg.Edit(locales.Trf("admin.muted", userId, userName))
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
	_, err = msg.Edit(locales.Trf("admin.unmuted", userId, userName))
	return err
}

func DmuteUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_dmute"))
		return err
	}
	if reason == "" {
		reason = locales.Tr("common.no_reason")
	}
	msg, _ := eOR(m, locales.Tr("admin.muting"))
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Mute: true})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.dmute_error"))
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(locales.Trf("admin.dmuted", userId, userName, reason))
	return err
}

func DkickUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_dkick"))
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
	_, err = msg.Edit(locales.Trf("admin.kicked", userId, userName, reason))
	return err
}

func DbanUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_dban"))
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
	_, err = msg.Edit(locales.Trf("admin.banned", userId, userName, reason))
	return err
}

func PromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_promote"))
		return err
	}
	if title == "" {
		title = "Admin"
	}
	msg, _ := eOR(m, locales.Tr("admin.promoting"))
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{
		Rights: &telegram.ChatAdminRights{
			ChangeInfo: true, DeleteMessages: true, InviteUsers: true,
			BanUsers: true, PinMessages: true,
		},
		Rank: title, IsAdmin: true,
	})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.promote_error"))
		return err
	}
	_, err = msg.Edit(locales.Trf("admin.promoted", userId, userName))
	return err
}

func FullPromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("admin.usage_fullpromote"))
		return err
	}
	if title == "" {
		title = "Super Admin"
	}
	msg, _ := eOR(m, locales.Tr("admin.promoting"))
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{
		Rights: &telegram.ChatAdminRights{
			ChangeInfo: true, DeleteMessages: true, InviteUsers: true,
			BanUsers: true, PinMessages: true, AddAdmins: true,
			ManageCall: true, ManageTopics: true, PostStories: true,
			EditStories: true, DeleteStories: true,
		},
		Rank: title, IsAdmin: true,
	})
	if err != nil {
		_, err := msg.Edit(locales.Tr("admin.promote_error"))
		return err
	}
	_, err = msg.Edit(locales.Trf("admin.promoted", userId, userName))
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
	_, err = msg.Edit(locales.Trf("admin.demoted", userId, userName))
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
	silent := strings.Contains(m.Args(), "silent")
	err = reply.Pin(&telegram.PinOptions{Silent: silent})
	if err != nil {
		_, err := eOR(m, locales.Tr("admin.pin_error"))
		return err
	}
	_, err = eOR(m, locales.Trf("admin.pinned", msgLink(reply)))
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
	_, err = eOR(m, locales.Trf("admin.unpinned", msgLink(reply)))
	return err
}

func zombiesCmd(m *telegram.NewMessage) error {
	args := m.Args()
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("admin.groups_only"))
		return err
	}
	perms, err := m.Client.GetChatMember(m.ChatID(), m.Sender.ID)
	if err != nil {
		_, err = eOR(m, locales.Tr("admin.get_permissions_error"))
		return err
	}
	if !perms.Rights.BanUsers {
		_, err = eOR(m, locales.Tr("admin.no_permissions"))
		return err
	}
	var deleted []int64
	msg, _ := eOR(m, locales.Tr("admin.zombies_searching"))
	members, _, err := m.Client.GetChatMembers(m.ChatID(), &telegram.ParticipantOptions{Limit: 500000})
	if err != nil {
		_, err = msg.Edit(locales.Tr("admin.get_members_error"))
		return err
	}
	for _, member := range members {
		if !member.User.Bot && member.User.Deleted {
			deleted = append(deleted, member.User.ID)
		}
	}
	if len(deleted) == 0 {
		_, err = msg.Edit(locales.Tr("admin.zombies_not_found"))
		return err
	}
	if strings.Contains(args, "clean") {
		success, failed := 0, 0
		for _, id := range deleted {
			if _, err := m.Client.KickParticipant(m.ChatID(), id); err != nil {
				failed++
			} else {
				success++
			}
		}
		_, err = msg.Edit(locales.Trf("admin.zombies_cleaned", success, failed))
		return err
	}
	_, err = msg.Edit(locales.Trf("admin.zombies_found", len(deleted)))
	return err
}

func PurgeCmd(m *telegram.NewMessage) error {
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_purge"))
		return err
	}
	var msgIDs []int32
	for id := reply.ID; id < m.ID; id++ {
		msgIDs = append(msgIDs, int32(id))
	}
	if len(msgIDs) > 0 {
		m.Client.DeleteMessages(m.ChatID(), msgIDs)
	}
	go func() {
		msg, _ := eOR(m, locales.Trf("admin.purged", len(msgIDs)))
		if msg != nil {
			<-time.After(2 * time.Second)
			msg.Delete()
		}
	}()
	return nil
}

func DeleteCmd(m *telegram.NewMessage) error {
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_del"))
		return err
	}
	_, err = m.Client.DeleteMessages(m.ChatID(), []int32{int32(reply.ID)})
	if err == nil {
		m.Delete()
	}
	return err
}

func PurgeFromCmd(m *telegram.NewMessage) error {
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_purgefrom"))
		return err
	}
	purgeFromMap.Lock()
	purgeFromMap.m[m.ChatID()] = reply.ID
	purgeFromMap.Unlock()
	_, err = eOR(m, locales.Tr("admin.purgefrom_set"))
	return err
}

func PurgeToCmd(m *telegram.NewMessage) error {
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, locales.Tr("admin.usage_purgeto"))
		return err
	}
	purgeFromMap.RLock()
	startID, ok := purgeFromMap.m[m.ChatID()]
	purgeFromMap.RUnlock()
	if !ok {
		_, err := eOR(m, locales.Tr("admin.purgeto_no_start"))
		return err
	}
	if startID >= reply.ID {
		_, err := eOR(m, locales.Tr("admin.purgeto_wrong_order"))
		return err
	}
	var msgIDs []int32
	for id := startID; id <= reply.ID; id++ {
		msgIDs = append(msgIDs, int32(id))
	}
	if len(msgIDs) > 0 {
		m.Client.DeleteMessages(m.ChatID(), msgIDs)
	}
	purgeFromMap.Lock()
	delete(purgeFromMap.m, m.ChatID())
	purgeFromMap.Unlock()
	_, err = eOR(m, locales.Trf("admin.purged", len(msgIDs)))
	return err
}

func loadAdminModule() {
	handlers := []*Handler{
		{Command: "ban", Func: BanUser, Description: locales.Tr("desc.ban"), ModuleName: "Admin"},
		{Command: "unban", Func: UnbanUser, Description: locales.Tr("desc.unban"), ModuleName: "Admin"},
		{Command: "kick", Func: KickUser, Description: locales.Tr("desc.kick"), ModuleName: "Admin"},
		{Command: "mute", Func: MuteUser, Description: locales.Tr("desc.mute"), ModuleName: "Admin"},
		{Command: "unmute", Func: UnmuteUser, Description: locales.Tr("desc.unmute"), ModuleName: "Admin"},
		{Command: "dmute", Func: DmuteUser, Description: locales.Tr("desc.dmute"), ModuleName: "Admin"},
		{Command: "demote", Func: DemoteUser, Description: locales.Tr("desc.demote"), ModuleName: "Admin"},
		{Command: "dkick", Func: DkickUser, Description: locales.Tr("desc.dkick"), ModuleName: "Admin"},
		{Command: "dban", Func: DbanUser, Description: locales.Tr("desc.dban"), ModuleName: "Admin"},
		{Command: "promote", Func: PromoteUser, Description: locales.Tr("desc.promote"), ModuleName: "Admin"},
		{Command: "fullpromote", Func: FullPromoteUser, Description: locales.Tr("desc.fullpromote"), ModuleName: "Admin"},
		{Command: "pin", Func: PinMessage, Description: locales.Tr("desc.pin"), ModuleName: "Admin"},
		{Command: "unpin", Func: UnpinMessage, Description: locales.Tr("desc.unpin"), ModuleName: "Admin"},
		{Command: "zombies", Func: zombiesCmd, Description: locales.Tr("desc.zombies"), ModuleName: "Admin"},
		{Command: "purge", Func: PurgeCmd, Description: locales.Tr("desc.purge"), ModuleName: "Admin"},
		{Command: "del", Func: DeleteCmd, Description: locales.Tr("desc.del"), ModuleName: "Admin"},
		{Command: "purgefrom", Func: PurgeFromCmd, Description: locales.Tr("desc.purgefrom"), ModuleName: "Admin"},
		{Command: "purgeto", Func: PurgeToCmd, Description: locales.Tr("desc.purgeto"), ModuleName: "Admin"},
	}
	AddHandlers(handlers, client)
}
