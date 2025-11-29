package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"bytes"
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func ping(ip string) (string, error) {
	out, err := utils.RunCommand(fmt.Sprintf("ping -c 1 -W 1 %s", ip))
	if err != nil {
		return "", err
	}

	if out != "" {
		idx := bytes.Index([]byte(out), []byte("time="))
		if idx >= 0 {
			start := idx + 5
			end := bytes.Index([]byte(out[start:]), []byte(" ms"))
			if end > 0 {
				return out[start : start+end], nil
			}
		}
	}
	return "timeout", nil
}

func DCPingHandler(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("ping.dc_pinging"))

	dcs := map[string]string{
		"DC1 (MIA, Miami FL, USA)": "149.154.175.53",
		"DC2 (AMS, Amsterdam, NL)": "149.154.167.51",
		"DC3 (MIA, Miami FL, USA)": "149.154.175.100",
		"DC4 (AMS, Amsterdam, NL)": "149.154.167.91",
		"DC5 (SIN, Singapore, SG)": "91.108.56.130",
	}

	response := locales.Tr("ping.dc_header") + "\n"
	for dcName, dcIP := range dcs {
		pingTime, err := ping(dcIP)
		if err != nil {
			response += fmt.Sprintf(locales.Tr("ping.dc_failed"), dcName) + "\n"
		} else {
			response += fmt.Sprintf(locales.Tr("ping.dc_entry"), dcName, pingTime) + "\n"
		}
		time.Sleep(100 * time.Millisecond)
	}

	_, err := msg.Edit(response)
	return err
}

func PingHandler(m *telegram.NewMessage) error {
	msgTime := m.OriginalUpdate.(*telegram.MessageObj).Date
	duration := time.Since(time.Unix(int64(msgTime), 0))

	msg, _ := eOR(m, locales.Tr("ping.pinging"))
	_, err := msg.Edit(fmt.Sprintf(locales.Tr("ping.result"), duration.Milliseconds(), time.Since(startTime).Truncate(time.Second)))
	return err
}

func Alive(m *telegram.NewMessage) error {
	uptime := time.Since(startTime).String()
	goVersion := runtime.Version()
	gogramVersion := telegram.Version

	message := fmt.Sprintf(locales.Tr("alive.message"),
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

func LoadSystemModule(c *telegram.Client) {
	handlers := []*Handler{
		{ModuleName: "System", Command: "ping", Description: "Ping the userbot", Func: PingHandler},
		{ModuleName: "System", Command: "dcping", Description: "Ping all data centers", Func: DCPingHandler},
		{ModuleName: "System", Command: "alive", Description: "Check if the bot is alive", Func: Alive},
	}
	AddHandlers(handlers, c)
}
