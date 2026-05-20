package modules

import (
	"NovaUserbot/locales"
	"context"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Alive", loadAliveModule)
}

func Alive(m *telegram.NewMessage) error {
	uptime := time.Since(startTime).String()
	goVersion := runtime.Version()
	gogramVersion := telegram.Version

	message := locales.Trf("alive.message",
		client.Me().FirstName+" "+client.Me().LastName, client.Me().ID,
		GetSudoersCount(), goVersion, gogramVersion, uptime,
	)
	aliveimage, err := Db.Get(context.Background(), "ALIVE_IMAGE").Result()
	if err == nil {
		_, err = eOR(m, message, &telegram.SendOptions{ParseMode: "HTML", Media: aliveimage})
	} else {
		_, err = eOR(m, message, &telegram.SendOptions{ParseMode: "HTML"})
	}
	return err
}

func loadAliveModule() {
	AddHandler(&Handler{
		Func:        Alive,
		Command:     "alive",
		Description: "Check if the bot is alive",
		ModuleName:  "Alive",
	}, client)
}
