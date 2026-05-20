package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Dev Tools", loadDevModule)
}

func ShellHandle(m *telegram.NewMessage) error {
	cmd := m.Args()
	if cmd == "" {
		eOR(m, locales.Tr("dev.no_command"))
		return nil
	}
	msg, _ := eOR(m, locales.Tr("dev.running"))
	out, err := utils.RunCommand(cmd)
	if err == nil && out == "" {
		_, err = msg.Edit(locales.Trf("dev.shell_no_output", m.Args()))
		return err
	}
	if len(out) > 4095 {
		tmpFile, err := os.CreateTemp("", "shell_output_*.txt")
		if err != nil {
			msg.Edit(locales.Trf("dev.shell_error", m.Args(), err.Error()))
			return err
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Write([]byte(out))
		tmpFile.Close()
		m.Client.SendMedia(m.ChatID, tmpFile.Name(), &telegram.MediaOptions{Caption: "Output"})
		return nil
	}
	_, err = msg.Edit(locales.Trf("dev.shell_result", m.Args(), out))
	return err
}

const boilerCodeForEval = `
package main

import "fmt"
import "github.com/amarnathcjd/gogram/telegram"
import "encoding/json"

var msg_id int32 = %d

var client *telegram.Client
var message *telegram.NewMessage
var m *telegram.NewMessage
var p func(...any) = func(a ...any) {
    for i, v := range a {
        valueType := reflect.TypeOf(v)
        if valueType.Kind() == reflect.Struct || (valueType.Kind() == reflect.Ptr && valueType.Elem().Kind() == reflect.Struct) {
            jsonData, err := json.MarshalIndent(v, "", "  ")
            if err != nil {
                a[i] = "Error marshalling struct"
            } else {
                a[i] = string(jsonData)
            }
        }
    }
    for _, v := range a {
        fmt.Println(v)
    }
}
var r *telegram.NewMessage
` + "var msg = `%s`\nvar snd = `%s`\nvar cht = `%s`\nvar chn = `%s`\nvar cch = `%s`" + `

func evalCode() {
	%s
}

func main() {
	var msg_o *telegram.MessageObj
	var snd_o *telegram.UserObj
	var cht_o *telegram.ChatObj
	var chn_o *telegram.Channel
	json.Unmarshal([]byte(msg), &msg_o)
	json.Unmarshal([]byte(snd), &snd_o)
	json.Unmarshal([]byte(cht), &cht_o)
	json.Unmarshal([]byte(chn), &chn_o)
	client, _ = telegram.NewClient(telegram.ClientConfig{
		StringSession: "%s",
		MemorySession: true,
	})

	client.Cache.ImportJSON([]byte(cch))
	client.Conn()

	x := []telegram.User{}
	y := []telegram.Chat{}
	x = append(x, snd_o)
	if chn_o != nil {
		y = append(y, chn_o)
	}
	if cht_o != nil {
		y = append(y, cht_o)
	}
	client.Cache.UpdatePeersToCache(x, y)
	idx := 0
	if cht_o != nil {
		idx = int(cht_o.ID)
	}
	if chn_o != nil {
		idx = int(chn_o.ID)
	}
	if snd_o != nil && idx == 0 {
		idx = int(snd_o.ID)
	}

	messageX, err := client.GetMessages(idx, &telegram.SearchOption{
		IDs: int(msg_id),
	})

	if err != nil {
		fmt.Println(err)
	}

	message = &messageX[0]
	m = message
	r, _ = message.GetReplyMessage()
	evalCode()
}

func packMessage(c *telegram.Client, message telegram.Message, sender *telegram.UserObj, channel *telegram.Channel, chat *telegram.ChatObj) *telegram.NewMessage {
	var (
		m = &telegram.NewMessage{}
	)
	switch message := message.(type) {
	case *telegram.MessageObj:
		m.ID = message.ID
		m.OriginalUpdate = message
		m.Message = message
		m.Client = c
	default:
		return nil
	}
	m.Sender = sender
	m.Chat = chat
	m.Channel = channel
	if m.Channel != nil && (m.Sender.ID == m.Channel.ID) {
		m.SenderChat = channel
	} else {
		m.SenderChat = &telegram.Channel{}
	}
	m.Peer, _ = c.GetSendablePeer(message.(*telegram.MessageObj).PeerID)
	return m
}
`

func EvalHandle(m *telegram.NewMessage) error {
	code := m.Args()
	if code == "" {
		_, err := eOR(m, locales.Tr("dev.no_code"))
		return err
	}
	msg, err := eOR(m, locales.Tr("dev.running"))
	if err != nil {
		return err
	}
	out, err := performEval(code, m)
	if err != nil {
		_, err = msg.Edit(locales.Trf("dev.error", err.Error()))
		return err
	}
	if len(out) > 3000 {
		tmpFile, err := os.CreateTemp("", "eval_output_*.txt")
		if err != nil {
			msg.Edit(locales.Trf("dev.error", err.Error()))
			return err
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Write([]byte(out))
		tmpFile.Close()
		m.Client.SendMedia(m.ChatID(), tmpFile.Name(), &telegram.MediaOptions{
			Caption: fmt.Sprintf("<b>Eval Code:</b> <pre lang='go'>%s</pre>", code),
		})
		msg.Delete()
		return nil
	}
	_, err = msg.Edit(locales.Trf("dev.eval_result", code, out))
	return err
}

func JsonHandle(m *telegram.NewMessage) error {
	var jsonString []byte
	if !m.IsReply() {
		switch {
		case strings.Contains(m.Args(), "-s"):
			jsonString, _ = json.MarshalIndent(m.Sender, "", "  ")
		case strings.Contains(m.Args(), "-m"):
			jsonString, _ = json.MarshalIndent(m.Media(), "", "  ")
		case strings.Contains(m.Args(), "-c"):
			jsonString, _ = json.MarshalIndent(m.Channel, "", "  ")
		default:
			jsonString, _ = json.MarshalIndent(m.OriginalUpdate, "", "  ")
		}
	} else {
		r, err := m.GetReplyMessage()
		if err != nil {
			eOR(m, locales.Trf("dev.error", err.Error()))
			return nil
		}
		switch {
		case strings.Contains(m.Args(), "-s"):
			jsonString, _ = json.MarshalIndent(r.Sender, "", "  ")
		case strings.Contains(m.Args(), "-m"):
			jsonString, _ = json.MarshalIndent(r.Media(), "", "  ")
		case strings.Contains(m.Args(), "-c"):
			jsonString, _ = json.MarshalIndent(r.Channel, "", "  ")
		default:
			jsonString, _ = json.MarshalIndent(r.OriginalUpdate, "", "  ")
		}
	}
	dataFieldRegex := regexp.MustCompile(`"Data": "([a-zA-Z0-9+/]+={0,2})"`)
	for _, v := range dataFieldRegex.FindAllStringSubmatch(string(jsonString), -1) {
		decoded, err := base64.StdEncoding.DecodeString(v[1])
		if err != nil {
			continue
		}
		jsonString = []byte(strings.ReplaceAll(string(jsonString), v[0], `"Data": "`+string(decoded)+`"`))
	}

	if len(jsonString) > 4095 {
		tmpFile, err := os.CreateTemp("", "json_output_*.json")
		if err != nil {
			eOR(m, err.Error())
			return err
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Write(jsonString)
		tmpFile.Close()
		m.ReplyMedia(tmpFile.Name(), &telegram.MediaOptions{Caption: "Message JSON"})
		m.Delete()
	} else {
		eOR(m, "<pre lang='json'>"+string(jsonString)+"</pre>")
	}
	return nil
}

func performEval(code string, m *telegram.NewMessage) (string, error) {
	msgB, _ := json.Marshal(m.Message)
	sndB, _ := json.Marshal(m.Sender)
	cntB, _ := json.Marshal(m.Chat)
	chnB, _ := json.Marshal(m.Channel)
	cacheB, _ := m.Client.Cache.ExportJSON()

	codeFile := fmt.Sprintf(boilerCodeForEval, m.ID, msgB, sndB, cntB, chnB, cacheB, code, m.Client.ExportSession())
	tmpDir, err := os.MkdirTemp("", "eval")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFileName := filepath.Join(tmpDir, "eval.go")
	if err := os.WriteFile(tmpFileName, []byte(codeFile), 0644); err != nil {
		return "", fmt.Errorf("error writing file: %v", err)
	}

	goImport, err := utils.RunCommand("goimports -w " + tmpFileName)
	if err != nil {
		return "", fmt.Errorf("goimports error: %v\n%s", err, goImport)
	}
	stdout, err := utils.RunCommand("go run " + tmpFileName)
	if err != nil {
		return "", fmt.Errorf("error: %v\n%s", err, stdout)
	}
	return stdout, nil
}

func loadDevModule() {
	handlers := []*Handler{
		{ModuleName: "Dev Tools", Command: "sh", Description: "Run shell commands", Func: ShellHandle, DisAllowSudos: true},
		{ModuleName: "Dev Tools", Command: "eval", Description: "Eval Go code", Func: EvalHandle, DisAllowSudos: true},
		{ModuleName: "Dev Tools", Command: "json", Description: "Get JSON of a message", Func: JsonHandle},
	}
	AddHandlers(handlers, client)
}
