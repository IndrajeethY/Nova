package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetVar(m *telegram.NewMessage) error {
	args := strings.SplitN(m.Args(), " ", 2)
	if len(args) < 2 {
		_, err := eOR(m, locales.Tr("database.usage_setvar"))
		return err
	}

	key := strings.ToUpper(strings.TrimSpace(args[0]))
	value := strings.TrimSpace(args[1])

	if key == "" || value == "" {
		_, err := eOR(m, locales.Tr("database.key_value_required"))
		return err
	}

	if err := db.Set(key, value); err != nil {
		_, err = eOR(m, fmt.Sprintf(locales.Tr("database.set_error"), err.Error()))
		return err
	}

	_, err := eOR(m, fmt.Sprintf(locales.Tr("database.set_success"), key, value))
	return err
}

func GetVar(m *telegram.NewMessage) error {
	key := strings.TrimSpace(m.Args())
	if key == "" {
		_, err := eOR(m, locales.Tr("database.usage_getvar"))
		return err
	}

	value := db.Get(strings.ToUpper(key))
	if value == "" {
		_, err := eOR(m, locales.Tr("database.get_not_found"))
		return err
	}

	_, err := eOR(m, fmt.Sprintf(locales.Tr("database.get_result"), strings.ToUpper(key), value))
	return err
}

func DelVar(m *telegram.NewMessage) error {
	key := strings.TrimSpace(m.Args())
	if key == "" {
		_, err := eOR(m, locales.Tr("database.usage_delvar"))
		return err
	}

	upperKey := strings.ToUpper(key)
	if !db.Exists(upperKey) {
		_, err := eOR(m, locales.Tr("database.del_not_found"))
		return err
	}

	if err := db.Del(upperKey); err != nil {
		_, err = eOR(m, fmt.Sprintf(locales.Tr("database.del_error"), err.Error()))
		return err
	}

	_, err := eOR(m, fmt.Sprintf(locales.Tr("database.del_success"), upperKey))
	return err
}

func ListVars(m *telegram.NewMessage) error {
	keys, err := db.Keys("*")
	if err != nil {
		_, err = eOR(m, locales.Tr("database.fetch_error"))
		return err
	}

	if len(keys) == 0 {
		_, err = eOR(m, locales.Tr("database.list_empty"))
		return err
	}

	var entries []string
	for _, key := range keys {
		entries = append(entries, fmt.Sprintf(locales.Tr("database.list_entry"), key))
	}

	msg := fmt.Sprintf(locales.Tr("database.list_header"), len(keys)) + "\n\n" + strings.Join(entries, "\n")
	_, err = eOR(m, msg, telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func DelAllVars(m *telegram.NewMessage) error {
	if m.Args() != "confirm" {
		_, err := eOR(m, locales.Tr("database.del_all_warning"))
		return err
	}

	if err := db.FlushAll(); err != nil {
		_, err = eOR(m, locales.Tr("database.del_all_error"))
		return err
	}

	_, err := eOR(m, locales.Tr("database.del_all_success"))
	return err
}

func LoadDbCmds(c *telegram.Client) {
	handlers := []*Handler{
		{Func: SetVar, Command: "setvar", Description: "Set a database variable", ModuleName: "Database"},
		{Func: GetVar, Command: "getvar", Description: "Get a database variable", ModuleName: "Database"},
		{Func: DelVar, Command: "delvar", Description: "Delete a database variable", ModuleName: "Database"},
		{Func: ListVars, Command: "vars", Description: "List all database variables", ModuleName: "Database"},
		{Func: DelAllVars, Command: "delallvars", Description: "Delete all variables (requires confirm)", ModuleName: "Database"},
	}
	AddHandlers(handlers, c)
}
