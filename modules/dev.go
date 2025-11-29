package modules

import (
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

func ShellHandle(m *telegram.NewMessage) error {
	cmd := m.Args()
	if cmd == "" {
		eOR(m, "No command provided")
		return nil
	}
	msg, _ := eOR(m, "<code>Running...</code>")
	out, err := utils.RunCommand(cmd)
	if err == nil && out == "" {
		_, err = msg.Edit(fmt.Sprintf("<b>CMD:</b> <pre lang='bash'>%s</pre>\n<b>OUTPUT:</b> <pre lang='bash'>No Output</pre>", m.Args()))
		return err
	}
	if len(out) > 4095 {
		tmpFile, err := os.Create("tmp/out.txt")
		defer os.Remove("tmp/out.txt")
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>CMD:</b> <pre lang='shell'>%s</pre>\n<b>ERROR:</b> <pre lang='shell'>%s</pre>", m.Args(), err.Error()))
		}
		_, err = tmpFile.Write([]byte(out))
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>CMD:</b> <pre lang='shell'>%s</pre>\n<b>ERROR:</b> <pre lang='shell'>%s</pre>", m.Args(), err.Error()))
		}
		_, err = m.Client.SendMedia(m.ChatID, tmpFile.Name(), &telegram.MediaOptions{Caption: "Output"})
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>CMD:</b> <pre lang='shell'>%s</pre>\n<b>ERROR:</b> <pre lang='shell'>%s</pre>", m.Args(), err.Error()))
		}
		return nil
	}
	_, err = msg.Edit(fmt.Sprintf("<b>CMD:</b> <pre lang='shell'>%s</pre>\n<b>OUTPUT:</b> <pre lang='shell'>%s</pre>", m.Args(), out))
	return err
}

const boiler_code_for_eval = `
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
		_, err := eOR(m, "<code>No code provided</code>")
		return err
	}
	msg, err := eOR(m, "<code>Running...</code>")
	if err != nil {
		return err
	}
	out, err := perfomEval(code, m)
	if err != nil {
		_, err = msg.Edit(fmt.Sprintf("<b>ERROR:</b> <pre lang='go'>%s</pre>", err.Error()))
		return err
	}
	if len(out) > 3000 {
		tmpFile, err := os.Create("out.txt")
		defer os.Remove(tmpFile.Name())
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>ERROR:</b> <pre lang='go'>%s</pre>", err.Error()))
			return err
		}
		_, err = tmpFile.Write([]byte(out))
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>ERROR:</b> <pre lang='go'>%s</pre>", err.Error()))
			return err
		}
		_, err = m.Client.SendMedia(m.ChatID(), tmpFile.Name(), &telegram.MediaOptions{Caption: fmt.Sprintf("<b>Eval Code:</b> <pre lang='go'>%s</pre>", code)})
		if err != nil {
			msg.Edit(fmt.Sprintf("<b>ERROR:</b> <pre lang='go'>%s</pre>", err.Error()))
			return err
		}
		msg.Delete()
		return nil
	}
	_, err = msg.Edit(fmt.Sprintf("<b>Eval Code:</b> <pre lang='go'>%s</pre>\n<b>Output:</b> <pre lang='go'>%s</pre>", code, out))
	return err
}

func JsonHandle(m *telegram.NewMessage) error {
	var jsonString []byte
	if !m.IsReply() {
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(m.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(m.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(m.Channel, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(m.OriginalUpdate, "", "  ")
		}
	} else {
		r, err := m.GetReplyMessage()
		if err != nil {
			m.Reply("<code>Error:</code> <b>" + err.Error() + "</b>")
			return nil
		}
		if strings.Contains(m.Args(), "-s") {
			jsonString, _ = json.MarshalIndent(r.Sender, "", "  ")
		} else if strings.Contains(m.Args(), "-m") {
			jsonString, _ = json.MarshalIndent(r.Media(), "", "  ")
		} else if strings.Contains(m.Args(), "-c") {
			jsonString, _ = json.MarshalIndent(r.Channel, "", "  ")
		} else {
			jsonString, _ = json.MarshalIndent(r.OriginalUpdate, "", "  ")
		}
	}
	dataFieldRegex := regexp.MustCompile(`"Data": "([a-zA-Z0-9+/]+={0,2})"`)
	dataFields := dataFieldRegex.FindAllStringSubmatch(string(jsonString), -1)
	for _, v := range dataFields {
		decoded, err := base64.StdEncoding.DecodeString(v[1])
		if err != nil {
			_, err = eOR(m, "Error: "+err.Error())
			return err
		}
		jsonString = []byte(strings.ReplaceAll(string(jsonString), v[0], `"Data": "`+string(decoded)+`"`))
	}

	if len(jsonString) > 4095 {
		defer os.Remove("message.json")
		tmpFile, err := os.Create("message.json")
		if err != nil {
			_, err = eOR(m, "Error: "+err.Error())
			return err
		}

		_, err = tmpFile.Write(jsonString)
		if err != nil {
			_, err = eOR(m, "Error: "+err.Error())
			return err
		}

		_, err = m.ReplyMedia(tmpFile.Name(), &telegram.MediaOptions{Caption: "Message JSON"})
		if err != nil {
			_, err = eOR(m, "Error: "+err.Error())
			return err
		}
		m.Delete()
	} else {
		_, err := eOR(m, "<pre lang='json'>"+string(jsonString)+"</pre>")
		return err
	}

	return nil
}

func perfomEval(code string, m *telegram.NewMessage) (string, error) {
	msg_b, _ := json.Marshal(m.Message)
	snd_b, _ := json.Marshal(m.Sender)
	cnt_b, _ := json.Marshal(m.Chat)
	chn_b, _ := json.Marshal(m.Channel)
	cache_b, _ := m.Client.Cache.ExportJSON()

	code_file := fmt.Sprintf(boiler_code_for_eval, m.ID, msg_b, snd_b, cnt_b, chn_b, cache_b, code, m.Client.ExportSession())
	tmpDir, err := os.MkdirTemp("", "eval")
	if err != nil {
		return "", fmt.Errorf("error creating temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFileName := filepath.Join(tmpDir, "eval.go")
	err = os.WriteFile(tmpFileName, []byte(code_file), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing file: %v", err)
	}

	return runGoCode(tmpFileName)
}

func runGoCode(fileName string) (string, error) {
	goImport, err := utils.RunCommand("goimports -w " + fileName)
	if err != nil {
		return "", fmt.Errorf("error running goimports: %v\nStderr: %s", err, goImport)
	}
	stdout, err := utils.RunCommand("go run " + fileName)
	if err != nil {
		return "", fmt.Errorf("error: %v\n%s", err, stdout)
	}
	return stdout, nil
}

func LoadShellHandler(c *telegram.Client) {
	handlers := []*Handler{
		{
			ModuleName:    "Dev Tools",
			Command:       "sh",
			Description:   "Used to run shell commands",
			Func:          ShellHandle,
			DisAllowSudos: true,
		},
		{
			ModuleName:    "Dev Tools",
			Command:       "eval",
			Description:   "Eval Go code",
			Func:          EvalHandle,
			DisAllowSudos: true,
		},
		{
			ModuleName:  "Dev Tools",
			Command:     "json",
			Description: "Get JSON of the message",
			Func:        JsonHandle,
		},
	}
	AddHandlers(handlers, c)
}
