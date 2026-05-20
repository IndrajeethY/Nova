package modules

import (
	"NovaUserbot/utils"
	"context"
	"fmt"
	"log"

	"github.com/amarnathcjd/gogram/telegram"
)

func AddSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .addsudo &lt;user_id&gt; or reply to  a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Adding sudo...</code>")
	if utils.IsIn64Array(sudoers, userId) {
		_, err := msg.Edit("<code>User is already a sudo</code>")
		return err
	}
	err := Db.SAdd(context.Background(), "SUDOS", userId).Err()
	if err != nil {
		_, err := msg.Edit("<code>Error adding sudo</code>")
		return err
	}

	sudoers = append(sudoers, userId)
	_, err = msg.Edit(fmt.Sprintf("<b>Sudo added <a href='tg://user?id=%d'>%s</a> as sudo</b>", userId, userName))
	return err
}

func DelSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "<code>Usage: .delsudo &lt;user_id&gt; or reply to  a user</code>")
		return err
	}
	msg, _ := eOR(m, "<code>Deleting sudo...</code>")
	if !utils.IsIn64Array(sudoers, userId) {
		_, err := msg.Edit("<code>User is not a sudo</code>")
		return err
	}
	err := Db.SRem(context.Background(), "SUDOS", userId).Err()
	if err != nil {
		log.Println("Error removing sudo:", err)
		_, err := msg.Edit("<code>Error deleting sudo</code>")
		return err
	}
	sudoers = utils.RemoveFrom64Array(sudoers, userId)
	_, err = msg.Edit(fmt.Sprintf("<b>Sudo deleted <a href='tg://user?id=%d'>%s</a> as sudo</b>", userId, userName))
	return err
}

func ListSudo(m *telegram.NewMessage) error {
	sudos, err := Db.SMembers(context.Background(), "SUDOS").Result()
	if err != nil {
		_, err := eOR(m, "<code>Error fetching sudos</code>")
		return err
	}
	var formattedSudos string
	msg, _ := eOR(m, "<code>Fetching sudos...</code>")
	for _, sudo := range sudos {
		userId, userName := GetUserInfo(sudo)
		formattedSudos += fmt.Sprintf("<b>▸</b> <a href='tg://user?id=%d'>%s</a>\n", userId, userName)
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Total sudos:</b> <code>%d</code>\n\n%s", len(sudos), formattedSudos), telegram.SendOptions{ParseMode: "HTML"})
	return err
}
func LoadSudoModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "addsudo", Func: AddSudo, Description: "Add user as sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "delsudo", Func: DelSudo, Description: "Delete user from sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "listsudo", Func: ListSudo, Description: "List all sudos", ModuleName: "Sudoers"},
	}
	AddHandlers(handlers, c)
}
