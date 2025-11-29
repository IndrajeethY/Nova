package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"
)

func AddSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("sudo.usage_add"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("sudo.adding"))

	if utils.IsIn64Array(sudoers, userId) {
		_, err := msg.Edit(locales.Tr("sudo.already_sudo"))
		return err
	}

	if err := db.SAdd("SUDOS", userId); err != nil {
		_, err := msg.Edit(locales.Tr("sudo.add_error"))
		return err
	}

	sudoers = append(sudoers, userId)
	_, err := msg.Edit(fmt.Sprintf(locales.Tr("sudo.added"), userId, userName))
	return err
}

func DelSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("sudo.usage_del"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("sudo.deleting"))

	if !utils.IsIn64Array(sudoers, userId) {
		_, err := msg.Edit(locales.Tr("sudo.not_sudo"))
		return err
	}

	if err := db.SRem("SUDOS", userId); err != nil {
		_, err := msg.Edit(locales.Tr("sudo.del_error"))
		return err
	}

	sudoers = utils.RemoveFrom64Array(sudoers, userId)
	_, err := msg.Edit(fmt.Sprintf(locales.Tr("sudo.deleted"), userId, userName))
	return err
}

func ListSudo(m *telegram.NewMessage) error {
	sudos, err := db.SMembers("SUDOS")
	if err != nil {
		_, err := eOR(m, locales.Tr("sudo.fetch_error"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("sudo.fetching"))

	var entries string
	for _, sudo := range sudos {
		userId, userName := GetUserInfo(sudo)
		entries += fmt.Sprintf(locales.Tr("sudo.list_entry"), userId, userName) + "\n"
	}

	result := fmt.Sprintf(locales.Tr("sudo.list_header"), len(sudos)) + "\n\n" + entries
	_, err = msg.Edit(result, &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func LoadSudoModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "addsudo", Func: AddSudo, Description: "Add user as sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "delsudo", Func: DelSudo, Description: "Remove user from sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "listsudo", Func: ListSudo, Description: "List all sudos", ModuleName: "Sudoers"},
	}
	AddHandlers(handlers, c)
}
