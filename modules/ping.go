package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"bytes"
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Ping", loadPingModule)
}

func ping(ip string) (string, error) {
	out, err := utils.RunCommand(fmt.Sprintf("ping -c 1 -W 1 %s", ip))
	if err != nil {
		return "", err
	}
	if len(out) > 0 {
		start := bytes.Index([]byte(out), []byte("time=")) + 5
		end := bytes.Index([]byte(out[start:]), []byte(" ms"))
		if start > 0 && end > 0 {
			return out[start : start+end], nil
		}
	}
	return "timeout", nil
}

func DCPingHandler(m *telegram.NewMessage) error {
	msg, err := eOR(m, locales.Tr("ping.dc_pinging"))
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

	response := locales.Tr("ping.dc_header")
	for dcName, dcIP := range dcs {
		pingTime, err := ping(dcIP)
		if err != nil {
			response += fmt.Sprintf(locales.Tr("ping.dc_failed"), dcName)
		} else {
			response += fmt.Sprintf(locales.Tr("ping.dc_entry"), dcName, pingTime)
		}
		time.Sleep(100 * time.Millisecond)
	}

	_, err = msg.Edit(response)
	return err
}

func PingHandler(m *telegram.NewMessage) error {
	msgObj, ok := m.OriginalUpdate.(*telegram.MessageObj)
	if !ok {
		return nil
	}
	msgTime := msgObj.Date
	duration := time.Since(time.Unix(int64(msgTime), 0))
	msg, err := eOR(m, locales.Tr("ping.pinging"))
	if err != nil {
		return err
	}
	_, err = msg.Edit(locales.Trf("ping.result", duration.Milliseconds(), time.Since(startTime).Truncate(time.Second)))
	return err
}

func loadPingModule() {
	handlers := []*Handler{
		{ModuleName: "Ping", Command: "ping", Description: locales.Tr("desc.ping"), Func: PingHandler},
		{ModuleName: "Ping", Command: "dcping", Description: locales.Tr("desc.dcping"), Func: DCPingHandler},
	}
	AddHandlers(handlers, client)
}
