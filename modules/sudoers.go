package modules

import (
	"NovaUserbot/locales"
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Sudoers", loadSudoModule)
}

func AddSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("sudo.usage_add"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("sudo.adding"))
	if IsSudoer(userId) {
		_, err := msg.Edit(locales.Tr("sudo.already_sudo"))
		return err
	}
	err := Db.SAdd(context.Background(), "SUDOS", userId).Err()
	if err != nil {
		_, err := msg.Edit(locales.Tr("sudo.add_error"))
		return err
	}
	AddSudoer(userId)
	_, err = msg.Edit(locales.Trf("sudo.added", userId, userName))
	return err
}

func DelSudo(m *telegram.NewMessage) error {
	userId, userName := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("sudo.usage_del"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("sudo.deleting"))
	if !IsSudoer(userId) {
		_, err := msg.Edit(locales.Tr("sudo.not_sudo"))
		return err
	}
	err := Db.SRem(context.Background(), "SUDOS", userId).Err()
	if err != nil {
		log.Println("Error removing sudo:", err)
		_, err := msg.Edit(locales.Tr("sudo.del_error"))
		return err
	}
	RemoveSudoer(userId)
	_, err = msg.Edit(locales.Trf("sudo.deleted", userId, userName))
	return err
}

func ListSudo(m *telegram.NewMessage) error {
	sudos, err := Db.SMembers(context.Background(), "SUDOS").Result()
	if err != nil {
		_, err := eOR(m, locales.Tr("sudo.fetch_error"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("sudo.fetching"))
	var b strings.Builder
	for _, sudo := range sudos {
		userId, userName := GetUserInfo(sudo)
		fmt.Fprintf(&b, locales.Tr("sudo.list_entry"), userId, userName)
	}
	_, err = msg.Edit(locales.Trf("sudo.list_result", len(sudos), b.String()), &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func loadSudoModule() {
	handlers := []*Handler{
		{Command: "addsudo", Func: AddSudo, Description: "Add user as sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "delsudo", Func: DelSudo, Description: "Remove user from sudo", ModuleName: "Sudoers", DisAllowSudos: true},
		{Command: "listsudo", Func: ListSudo, Description: "List all sudos", ModuleName: "Sudoers"},
	}
	AddHandlers(handlers, client)
}
