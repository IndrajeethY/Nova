package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	chatbotPrompt = "Make sure to give only one word response. The image may contain alphabetical characters, emojis, math problems, or a country flag. If it contains words, just give the word; if it's a math problem, solve it and provide the result with proper sign; if it's an emoji, print it; if it's a country flag, print the country name."
	chatbotID     = 691070694
)

func init() {
	RegisterModule("ChatBot", loadChatBotModule)
}

func OnChatBotMessage(m *telegram.NewMessage) error {
	if m.Sender.ID != chatbotID {
		return nil
	}
	if m.Media() == nil || !strings.Contains(m.Text(), "minutes") {
		return nil
	}
	file, err := m.Client.DownloadMedia(m.Media())
	if err != nil {
		log.Error("Error downloading media:", err)
		return err
	}
	defer os.Remove(file)
	result, err := utils.ProcessGemini(file, chatbotPrompt)
	if err != nil {
		log.Error("Error processing image:", err)
		return err
	}
	if m.Message.ReplyMarkup != nil {
		markup, ok := m.Message.ReplyMarkup.(*telegram.ReplyInlineMarkup)
		if !ok {
			return nil
		}
		for _, row := range markup.Rows {
			for _, btns := range row.Buttons {
				btn, ok := btns.(*telegram.KeyboardButtonCallback)
				if !ok {
					continue
				}
				btnText := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(btn.Text, "\n", ""), "\r", ""))
				resultText := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(result, "\n", ""), "\r", ""))
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

func geminiAi(m *telegram.NewMessage) error {
	var image string
	args := m.Args()
	msg, _ := eOR(m, locales.Tr("chatbot.fetching"))
	if m.IsReply() {
		reply, _ := m.GetReplyMessage()
		if reply != nil {
			if reply.Media() != nil {
				image, _ = m.Client.DownloadMedia(reply.Media())
				defer os.Remove(image)
			}
			if reply.Text() != "" {
				args = reply.Text()
			}
		}
	}
	if args == "" {
		_, err := msg.Edit(locales.Tr("chatbot.no_query"))
		return err
	}
	result, err := utils.ProcessGemini(image, args)
	if err != nil {
		_, err = msg.Edit(locales.Tr("chatbot.error"))
		return err
	}
	_, err = msg.Edit(locales.Trf("chatbot.result", args, result), &telegram.SendOptions{ParseMode: "Markdown"})
	return err
}

func loadChatBotModule() {
	AddHandler(&Handler{
		ModuleName:  "ChatBot",
		Command:     "ai",
		Description: "Fetch response from Gemini AI",
		Func:        geminiAi,
	}, client)
	client.On("message", OnChatBotMessage)
}
