package modules

import (
	"NovaUserbot/utils"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	promt = "Make sure to give only one word response. The image may contain alphabetical characters, emojis, math problems, or a country flag. If it contains words, just give the word; if it's a math problem, solve it and provide the result with proper sign; if it's an emoji, print it; if it's a country flag, print the country name."
	cId   = 691070694
)

func OnChatBotMessage(m *telegram.NewMessage) error {
	if m.Sender.ID != cId {
		return nil
	}
	if m.Media() != nil {
		if !strings.Contains(m.Text(), "minutes") {
			return nil
		}
		file, err := m.Client.DownloadMedia(m.Media())
		if err != nil {
			log.Error("Error downloading media:", err)
			return err
		}
		result, err := utils.ProcessGemini(file, promt)
		if err != nil {
			log.Error("Error processing image and generating content:", err)
			return err
		}
		if m.Message.ReplyMarkup != nil {
			for _, row := range m.Message.ReplyMarkup.(*telegram.ReplyInlineMarkup).Rows {
				for _, btns := range row.Buttons {
					btn, ok := btns.(*telegram.KeyboardButtonCallback)
					if !ok {
						log.Error("Error casting button to KeyboardButtonCallback")
						continue
					}
					btnText := strings.TrimSpace(btn.Text)
					resultText := strings.TrimSpace(result)
					btnText = strings.ReplaceAll(btnText, "\n", "")
					btnText = strings.ReplaceAll(btnText, "\r", "")
					resultText = strings.ReplaceAll(resultText, "\n", "")
					resultText = strings.ReplaceAll(resultText, "\r", "")
					if strings.Contains(strings.ToLower(btnText), strings.ToLower(resultText)) {
						_, err = m.Click(btn.Data)
						return err
					}
				}
			}
		} else {
			_, err = m.Respond(result)
		}
		return err
	}
	return nil
}

func geminiAi(m *telegram.NewMessage) error {
	var image string
	args := m.Args()
	msg, _ := eOR(m, "<code>Fetching Response...</code>")
	if m.IsReply() {
		msg, _ := m.GetReplyMessage()
		if msg.Media() != nil {
			image, _ = m.Client.DownloadMedia(msg.Media())
			defer os.Remove(image)
		}
		if msg.Text() != "" {
			args = msg.Text()
		}

	}
	if args == "" {
		_, err := msg.Edit("<code>No Query Provided</code>")
		return err
	}
	result, err := utils.ProcessGemini(image, args)
	if err != nil {
		_, err = msg.Edit("<code>Error Fetching Data</code>")
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("**Query:** `%s`\n\n**Response:**\n%s", args, result), telegram.SendOptions{ParseMode: "Markdown"})
	return err
}

func LoadChatBotHandler(c *telegram.Client) {
	handlers := []*Handler{
		{
			ModuleName:  "ChatBot",
			Command:     "ai",
			Description: "Fetch response from Gemini AI",
			Func:        geminiAi,
		},
	}
	AddHandlers(handlers, c)
	c.On("message", OnChatBotMessage)
}
