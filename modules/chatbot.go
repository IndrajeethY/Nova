package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"fmt"
	"os"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	chatbotPrompt = "Make sure to give only one word response. The image may contain alphabetical characters, emojis, math problems, or a country flag."
	chatbotId     = 691070694
)

func OnChatBotMessage(m *telegram.NewMessage) error {
	if m.Sender.ID != chatbotId || m.Media() == nil {
		return nil
	}

	if !strings.Contains(m.Text(), "minutes") {
		return nil
	}

	file, err := m.Client.DownloadMedia(m.Media())
	if err != nil {
		logger.Error("Chatbot download error:", err)
		return err
	}

	result, err := utils.ProcessGemini(file, chatbotPrompt)
	if err != nil {
		logger.Error("Chatbot Gemini error:", err)
		return err
	}

	if m.Message.ReplyMarkup != nil {
		for _, row := range m.Message.ReplyMarkup.(*telegram.ReplyInlineMarkup).Rows {
			for _, btn := range row.Buttons {
				if cb, ok := btn.(*telegram.KeyboardButtonCallback); ok {
					btnText := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(cb.Text, "\n", ""), "\r", ""))
					resultText := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(result, "\n", ""), "\r", ""))
					if strings.Contains(strings.ToLower(btnText), strings.ToLower(resultText)) {
						m.Click(cb.Data)
						return nil
					}
				}
			}
		}
	}

	m.Respond(result)
	return nil
}

func geminiAi(m *telegram.NewMessage) error {
	var image string
	args := m.Args()

	msg, _ := eOR(m, locales.Tr("chatbot.fetching"))

	if m.IsReply() {
		reply, _ := m.GetReplyMessage()
		if reply.Media() != nil {
			image, _ = m.Client.DownloadMedia(reply.Media())
			defer os.Remove(image)
		}
		if reply.Text() != "" {
			args = reply.Text()
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

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("chatbot.result"), args, result), &telegram.SendOptions{ParseMode: "Markdown"})
	return err
}

func LoadChatBotHandler(c *telegram.Client) {
	handlers := []*Handler{
		{ModuleName: "ChatBot", Command: "ai", Description: "Query Gemini AI", Func: geminiAi},
	}
	AddHandlers(handlers, c)
	c.On("message", OnChatBotMessage)
}
