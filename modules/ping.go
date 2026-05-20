package modules

import (
	"NovaUserbot/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func ping(ip string) (string, error) {
	out, err := utils.RunCommand(fmt.Sprintf("ping -c 1 -W 1 %s", ip))
	if err != nil {
		return "", err
	}

	if len(out) > 0 && out != "" {
		start := bytes.Index([]byte(out), []byte("time=")) + 5
		end := bytes.Index([]byte(out[start:]), []byte(" ms"))
		if start > 0 && end > 0 {
			return out[start : start+end], nil
		}
	}

	return "timeout", nil
}

func DCPingHandler(m *telegram.NewMessage) error {
	msg, err := eOR(m, "<code>Pinging all DCs...</code>")
	if err != nil {
		return err
	}

	dcs := map[string]string{
		"DC1 (MIA, Miami FL, USA)": "149.154.175.53",
		"DC2 (AMS, Amsterdam, NL)": "149.154.167.51",
		"DC3 (MIA, Miami FL, USA)": "149.154.175.100",
		"DC4 (AMS, Amsterdam, NL)": "149.154.167.91",
		"DC5 (SIN, Singapore, SG)": "91.108.56.130",
	}

	response := "<b>Data Center Pings:</b>\n"
	for dcName, dcIP := range dcs {
		pingTime, err := ping(dcIP)
		if err != nil {
			response += fmt.Sprintf("<b>%s:</b> <code>Failed to ping</code>\n", dcName)
		} else {
			response += fmt.Sprintf("<b>%s:</b> <code>%s</code> ms\n", dcName, pingTime)
		}
		time.Sleep(100 * time.Millisecond)
	}

	_, err = msg.Edit(response)
	return err
}

func PingHandler(m *telegram.NewMessage) error {
	msgTime := m.OriginalUpdate.(*telegram.MessageObj).Date
	duration := time.Since(time.Unix(int64(msgTime), 0))
	msg, err := eOR(m, "<code>Pinging...</code>")
	if err != nil {
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Pong!</b>\n<b>Time taken:</b> <code>%v</code>ms\n<b>Uptime:</b> <code>%v</code>", duration.Milliseconds(), time.Since(startTime).Truncate(time.Second)))
	return err
}

func LoadPingHandler(c *telegram.Client) {
	handlers := []*Handler{
		{
			ModuleName:  "Ping Cmds",
			Command:     "ping",
			Description: "Ping the userbot",
			Func:        PingHandler,
		},
		{
			ModuleName:  "Ping Cmds",
			Command:     "dcping",
			Description: "Ping all data centers",
			Func:        DCPingHandler,
		},
	}
	AddHandlers(handlers, c)
}
