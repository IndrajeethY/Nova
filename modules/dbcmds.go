package modules

import (
	"context"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetKey(m *telegram.NewMessage) error {
	args := strings.Split(m.Args(), " ")
	if len(args) < 2 {
		_, err := eOR(m, "<code>Usage: .setkey &lt;key&gt; &lt;value&gt;</code>")
		return err
	}
	key := strings.ToUpper(args[0])
	value := strings.TrimSpace(strings.Join(args[1:], " "))
	err := Db.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		_, err = eOR(m, "<code>Error setting key</code>")
	}
	_, err = eOR(m, fmt.Sprintf("<b>Key set successfully:</b>\n<b>Key:</b> <code>%s</code>\n<b>Value:</b> <code>%v</code>", key, value))
	return err
}

func GetKey(m *telegram.NewMessage) error {
	key := m.Args()
	if key == "" {
		_, err := eOR(m, "<code>Usage: .getkey &lt;key&gt;</code>")
		return err
	}
	config, err := Db.Get(context.Background(), strings.ToUpper(key)).Result()
	if err != nil {
		_, err = eOR(m, "<code>Key not found</code>")
		return err
	}
	_, err = eOR(m, fmt.Sprintf("<b>Key:</b> <code>%s</code>\n<b>Value:</b> <code>%s</code>", key, config))
	return err
}

func DelKey(m *telegram.NewMessage) error {
	key := m.Args()
	if key == "" {
		_, err := eOR(m, "<code>Usage: .delkey &lt;key&gt;</code>")
		return err
	}
	err := Db.Del(context.Background(), strings.ToUpper(key)).Err()
	if err != nil {
		_, err = eOR(m, "<code>Key not found</code>")
		return err
	}
	_, err = eOR(m, fmt.Sprintf("<b>Key deleted successfully:</b><code>%s</code>", key))
	return err
}

func ListKeys(m *telegram.NewMessage) error {
	keys, err := Db.Keys(context.Background(), "*").Result()
	if err != nil {
		_, err = m.Edit("<code>Error fetching keys</code>")
		return err
	}
	var formattedKeys []string
	for _, key := range keys {
		formattedKeys = append(formattedKeys, fmt.Sprintf("<b>▸</b> <code>%s</code>", key))
	}
	_, err = eOR(m, fmt.Sprintf("<b>Total keys:</b> <code>%d</code>\n\n%s", len(keys), strings.Join(formattedKeys, "\n")), telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func DelAllKeys(m *telegram.NewMessage) error {
	err := Db.FlushAll(context.Background()).Err()
	if err != nil {
		_, err = eOR(m, "<code>Error deleting keys</code>")
		return err
	}
	_, err = eOR(m, "<b>All keys deleted successfully</b>")
	return err
}

func LoadDbCmds(c *telegram.Client) {
	handlers := []*Handler{
		{Func: SetKey, Command: "setkey", Description: "Set a key in the database", ModuleName: "Database Cmds"},
		{Func: GetKey, Command: "getkey", Description: "Get a key from the database", ModuleName: "Database Cmds"},
		{Func: DelKey, Command: "delkey", Description: "Delete a key from the database", ModuleName: "Database Cmds"},
		{Func: ListKeys, Command: "listkeys", Description: "List all keys in the database", ModuleName: "Database Cmds"},
		{Func: DelAllKeys, Command: "delallkeys", Description: "Delete all keys from the database", ModuleName: "Database Cmds"},
	}
	AddHandlers(handlers, c)
}
