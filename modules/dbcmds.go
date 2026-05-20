package modules

import (
	"NovaUserbot/locales"
	"context"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Database", loadDbModule)
}

func SetKey(m *telegram.NewMessage) error {
	args := strings.Split(m.Args(), " ")
	if len(args) < 2 {
		_, err := eOR(m, locales.Tr("database.usage_setkey"))
		return err
	}
	key := strings.ToUpper(args[0])
	value := strings.TrimSpace(strings.Join(args[1:], " "))
	err := Db.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		_, err = eOR(m, locales.Tr("database.set_error"))
		return err
	}
	_, err = eOR(m, locales.Trf("database.set_success", key, value))
	return err
}

func GetKey(m *telegram.NewMessage) error {
	key := m.Args()
	if key == "" {
		_, err := eOR(m, locales.Tr("database.usage_getkey"))
		return err
	}
	val, err := Db.Get(context.Background(), strings.ToUpper(key)).Result()
	if err != nil {
		_, err = eOR(m, locales.Tr("database.get_not_found"))
		return err
	}
	_, err = eOR(m, locales.Trf("database.get_result", key, val))
	return err
}

func DelKey(m *telegram.NewMessage) error {
	key := m.Args()
	if key == "" {
		_, err := eOR(m, locales.Tr("database.usage_delkey"))
		return err
	}
	err := Db.Del(context.Background(), strings.ToUpper(key)).Err()
	if err != nil {
		_, err = eOR(m, locales.Tr("database.del_not_found"))
		return err
	}
	_, err = eOR(m, locales.Trf("database.del_success", key))
	return err
}

func ListKeys(m *telegram.NewMessage) error {
	keys, err := Db.Keys(context.Background(), "*").Result()
	if err != nil {
		_, err = eOR(m, locales.Tr("database.fetch_error"))
		return err
	}
	var b strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&b, locales.Tr("database.list_entry"), key)
		b.WriteString("\n")
	}
	_, err = eOR(m, locales.Trf("database.list_header", len(keys), b.String()), &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func DelAllKeys(m *telegram.NewMessage) error {
	err := Db.FlushDB(context.Background()).Err()
	if err != nil {
		_, err = eOR(m, locales.Tr("database.del_all_error"))
		return err
	}
	_, err = eOR(m, locales.Tr("database.del_all_success"))
	return err
}

var configKeys = []struct {
	Key  string
	Desc string
}{
	{"ALIVE_IMAGE", "Image/media URL for .alive command"},
	{"CMD_HANDLER", "Command prefix (default: .)"},
	{"LOG_CHAT", "Chat ID for log messages"},
	{"TAG_LOGGER", "Chat ID for tag notifications"},
	{"PM_AI_PROMT", "Custom AI prompt for PM permit"},
	{"PM_BOT_ENABLED", "PM Bot toggle (true/false)"},
	{"PM_BOT_LOG_GROUP", "Forum group ID for PM Bot topics"},
}

func AvailableKeys(m *telegram.NewMessage) error {
	var b strings.Builder
	for _, k := range configKeys {
		fmt.Fprintf(&b, locales.Tr("database.keys_entry"), k.Key, k.Desc)
		b.WriteString("\n")
	}
	_, err := eOR(m, locales.Trf("database.keys_header", len(configKeys), b.String()), &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func loadDbModule() {
	handlers := []*Handler{
		{Func: SetKey, Command: "setkey", Description: "Set a key in the database", ModuleName: "Database"},
		{Func: GetKey, Command: "getkey", Description: "Get a key from the database", ModuleName: "Database"},
		{Func: DelKey, Command: "delkey", Description: "Delete a key from the database", ModuleName: "Database"},
		{Func: ListKeys, Command: "listkeys", Description: "List all keys in the database", ModuleName: "Database"},
		{Func: DelAllKeys, Command: "delallkeys", Description: "Delete all keys from the database", ModuleName: "Database"},
		{Func: AvailableKeys, Command: "keys", Description: "Show all available config keys", ModuleName: "Database"},
	}
	AddHandlers(handlers, client)
}
