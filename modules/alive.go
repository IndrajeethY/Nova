package modules

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func Alive(m *telegram.NewMessage) error {
	uptime := time.Since(startTime).String()
	goVersion := runtime.Version()
	gogramVersion := telegram.Version

	message := fmt.Sprintf(
		"<b>Nᴏᴠᴀ Is Aʟɪᴠᴇ!</b>\n\n"+
			"<b>Oᴡɴᴇʀ:</b> <code>%s</code> (<code>%d</code>)\n"+
			"<b>Sᴜᴅᴏs Cᴏᴜɴᴛ:</b> <code>%d</code>\n"+
			"<b>Gᴏ Vᴇʀsɪᴏɴ:</b> <code>%s</code>\n"+
			"<b>Gᴏɢʀᴀᴍ Vᴇʀsɪᴏɴ:</b> <code>%s</code>\n"+
			"<b>Uᴘᴛɪᴍᴇ:</b> <code>%s</code>",
		client.Me().FirstName+" "+client.Me().LastName, client.Me().ID, len(sudoers), goVersion, gogramVersion, uptime,
	)
	aliveimage, err := Db.Get(context.Background(), "ALIVE_IMAGE").Result()
	if err == nil {
		_, err = eOR(m, message, telegram.SendOptions{ParseMode: "HTML", Media: aliveimage})
	} else {
		_, err = eOR(m, message, telegram.SendOptions{ParseMode: "HTML", Media: "https://files.indrajeeth.in/nova.jpg"})
	}
	return err
}

func LoadAliveCmd(c *telegram.Client) {
	handler := &Handler{
		Func:        Alive,
		Command:     "alive",
		Description: "Check if the bot is alive",
		ModuleName:  "Alive Cmd",
	}
	AddHandlers([]*Handler{handler}, c)
}
