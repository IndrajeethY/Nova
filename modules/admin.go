package modules

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func BanUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .ban &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}
	msg, _ := eOR(m, "<code>Banning user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Ban: true})
	if err != nil {
		_, err := msg.Edit("<code>Error banning user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Banned <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s", userId, userName, reason))
	return err
}

func UnbanUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .unban &lt;user_id&gt; or reply to a user</code>")
		return err
	}

	msg, _ := eOR(m, "<code>Unbanning user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Unban: true})
	if err != nil {
		_, err := msg.Edit("<code>User is not banned</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Unbanned <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func KickUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .kick &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}
	msg, _ := eOR(m, "<code>Kicking user...</code>")
	_, err := m.Client.KickParticipant(m.ChatID(), userId)
	if err != nil {
		_, err := msg.Edit("<code>Error kicking user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Kicked <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s", userId, userName, reason))
	return err
}

func MuteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .mute &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Muting user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Mute: true})
	if err != nil {
		_, err := msg.Edit("<code>Error muting user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Muted <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func UnmuteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .unmute &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Unmuting user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Unmute: true})
	if err != nil {
		_, err := msg.Edit("<code>User is not muted</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Unmuted <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func DmuteUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .dmute &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}
	msg, _ := eOR(m, "<code>Demoting and muting user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Mute: true})
	if err != nil {
		_, err := msg.Edit("<code>Error demoting and muting user</code>")
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Muted <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s", userId, userName, reason))
	return err
}

func DkickUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .kick &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}
	msg, _ := eOR(m, "<code>Kicking user...</code>")
	_, err := m.Client.KickParticipant(m.ChatID(), userId)
	if err != nil {
		_, err := msg.Edit("<code>Error kicking user</code>")
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Kicked <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s", userId, userName, reason))
	return err
}

func DbanUser(m *telegram.NewMessage) error {
	userId, userName, reason := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .ban &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if reason == "" {
		reason = "No reason"
	}
	msg, _ := eOR(m, "<code>Banning user...</code>")
	_, err := m.Client.EditBanned(m.ChatID(), userId, &telegram.BannedOptions{Ban: true})
	if err != nil {
		_, err := msg.Edit("<code>Error banning user</code>")
		return err
	}
	if reply, _ := m.GetReplyMessage(); reply != nil {
		reply.Delete()
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Banned <a href='tg://user?id=%d'>%s</a></b>\n<b>Reason:</b> %s", userId, userName, reason))
	return err
}

func PromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .promote &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if title == "" {
		title = "Λ∂мιи"
	}
	msg, _ := eOR(m, "<code>Promoting user...</code>")
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{Rights: &telegram.ChatAdminRights{ChangeInfo: true, DeleteMessages: true, InviteUsers: true, BanUsers: true, PinMessages: true}, Rank: title, IsAdmin: true})
	if err != nil {
		_, err := msg.Edit("<code>Error promoting user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Promoted <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func FullPromoteUser(m *telegram.NewMessage) error {
	userId, userName, title := ExtractUserMsg(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .promote &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	if title == "" {
		title = "𝙎υρєя Λ∂мιи"
	}
	msg, _ := eOR(m, "<code>Promoting user...</code>")
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{Rights: &telegram.ChatAdminRights{ChangeInfo: true, DeleteMessages: true, InviteUsers: true, BanUsers: true, PinMessages: true, AddAdmins: true, ManageCall: true, ManageTopics: true, PostStories: true, EditStories: true, DeleteStories: true}, Rank: title, IsAdmin: true})
	if err != nil {
		_, err := msg.Edit("<code>Error promoting user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Promoted <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func DemoteUser(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .demote &lt;user_id&gt; or reply to a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Demoting user...</code>")
	_, err := m.Client.EditAdmin(m.ChatID(), userId, &telegram.AdminOptions{IsAdmin: false, Rights: &telegram.ChatAdminRights{}})
	if err != nil {
		_, err := msg.Edit("<code>Error demoting user</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Demoted <a href='tg://user?id=%d'>%s</a></b>", userId, userName))
	return err
}

func PinMessage(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, "This command can only be used in groups.")
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, "<code>Usage: .pin (reply to a message)</code>")
		return err
	}
	var silent bool
	if strings.Contains(m.Args(), "silent") {
		silent = true
	}
	err = reply.Pin(&telegram.PinOptions{Silent: silent})
	if err != nil {
		_, err := eOR(m, "<code>Error pinning message</code>")
		return err
	}
	_, err = eOR(m, fmt.Sprintf("<b>Pinned <a href='%s'>this</a> message</b>", msgLink(reply)))
	return err
}

func UnpinMessage(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, "This command can only be used in groups.")
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err := eOR(m, "<code>Usage: .unpin (reply to a message)</code>")
		return err
	}
	err = reply.Pin(&telegram.PinOptions{Unpin: true})
	if err != nil {
		_, err := eOR(m, "<code>Error pinning message</code>")
		return err
	}
	_, err = eOR(m, fmt.Sprintf("<b>Unpinned <a href='%s'>this</a> message</b>", msgLink(reply)))
	return err
}

func zombiesCmd(m *telegram.NewMessage) error {
	args := m.Args()
	if !m.IsGroup() {
		_, err := m.Edit("This command can only be used in groups.")
		return err
	}
	perms, err := m.Client.GetChatMember(m.ChatID(), m.Sender.ID)
	if err != nil {
		_, err = m.Edit("Failed to get your permissions.")
		return err
	}
	if !perms.Rights.BanUsers {
		_, err = m.Edit("You don't have enough permissions to use this command.")
		return err
	}
	deleted := []int64{}
	msg, _ := m.Edit("<code>Searching for zombies...</code>")
	members, _, err := m.Client.GetChatMembers(m.ChatID(), &telegram.ParticipantOptions{Limit: 500000})
	if err != nil {
		_, err = msg.Edit("Failed to get chat members.")
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
		_, err = msg.Edit("No Deleted accounts found.")
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
		_, err = msg.Edit(fmt.Sprintf("Cleaned %d zombies, %d failed.", success, failed))
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("Found %d zombies. Use <code>.zombies clean</code> to remove them.", len(deleted)))
	return err
}

func LoadAdminModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "ban", Func: BanUser, Description: "Ban a user from the chat. Usage: .ban <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "unban", Func: UnbanUser, Description: "Unban a user from the chat. Usage: .unban <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "kick", Func: KickUser, Description: "Kick a user from the chat. Usage: .kick <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "mute", Func: MuteUser, Description: "Mute a user in the chat. Usage: .mute <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "unmute", Func: UnmuteUser, Description: "Unmute a user in the chat. Usage: .unmute <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "dmute", Func: DmuteUser, Description: "Demote and mute a user in the chat. Usage: .dmute <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "demote", Func: DemoteUser, Description: "Demote a user from admin. Usage: .demote <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "dkick", Func: DkickUser, Description: "Demote and kick a user from the chat. Usage: .dkick <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "dban", Func: DbanUser, Description: "Demote and ban a user from the chat. Usage: .dban <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "promote", Func: PromoteUser, Description: "Promote a user to admin. Usage: .promote <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "fullpromote", Func: FullPromoteUser, Description: "Fully promote a user to super admin. Usage: .fullpromote <user_id> or reply to a user", ModuleName: "Admin"},
		{Command: "pin", Func: PinMessage, Description: "Pin a message in the chat. Usage: .pin (reply to a message)", ModuleName: "Admin"},
		{Command: "unpin", Func: UnpinMessage, Description: "Unpin a message in the chat. Usage: .unpin (reply to a message)", ModuleName: "Admin"},
		{Command: "zombies", Func: zombiesCmd, Description: "Find and clean deleted accounts in the chat", ModuleName: "Admin"},
	}
	AddHandlers(handlers, c)
}
