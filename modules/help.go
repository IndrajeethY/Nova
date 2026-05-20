package modules

import (
	"NovaUserbot/utils"
	"fmt"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

type Handler struct {
	Command       string
	Description   string
	ModuleName    string
	Func          any
	DisAllowSudos bool
}

var HelpMap = map[string][]Handler{}

var ModuleList []string

func LoadModulesOrder() {
	for mod := range HelpMap {
		ModuleList = append(ModuleList, mod)
	}
	sort.Strings(ModuleList)
}

func HelpInline(i *telegram.InlineQuery) error {
	b := i.Builder()
	if !utils.IsIn64Array(sudoers, i.Sender.ID) && i.Sender.ID != ubId {
		btn := telegram.ButtonBuilder{}
		b.Article("Not Allowed", "You are not allowed to use this bot", "Not Allowed", &telegram.ArticleOptions{ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL("Owner", "t.me/tamilvip007")).Build()})
		i.Answer(b.Results())
		return nil
	}
	b.Article("Help Menu", "Available Help Menu", "<b>Aᴠᴀɪʟᴀʙʟᴇ Hᴇʟᴘ Mᴏᴅᴜʟᴇs:</b>", &telegram.ArticleOptions{ReplyMarkup: PaginateHelp(0), ID: "help"})
	i.Answer(b.Results())
	return nil
}

func PaginateHelp(index int) *telegram.ReplyInlineMarkup {
	b := telegram.ButtonBuilder{}
	var tgbtns []telegram.KeyboardButton
	max := 6

	totalModules := len(ModuleList)

	start := index * max
	end := min(start+max, totalModules)

	for _, mod := range ModuleList[start:end] {
		tgbtns = append(tgbtns, b.Data(mod, fmt.Sprintf("help:%s:%d", strings.ReplaceAll(mod, " ", "_"), index)))
	}

	if index > 0 {
		tgbtns = append(tgbtns, b.Data("⬅ Back", fmt.Sprintf("help_page_%d", index-1)))
	}
	if end < totalModules {
		tgbtns = append(tgbtns, b.Data("Next ➡", fmt.Sprintf("help_page_%d", index+1)))
	}

	return telegram.NewKeyboard().NewGrid(4, 2, tgbtns...).Build()
}

func HelpCmd(m *telegram.NewMessage) error {
	results, _ := m.Client.InlineQuery(tbotId, &telegram.InlineOptions{Query: "help"})
	res := results.Results[0].(*telegram.BotInlineResultObj)
	defer m.Delete()
	chat, _ := m.Client.GetSendablePeer(m.ChatID())
	_, err := m.Client.MessagesSendInlineBotResult(&telegram.MessagesSendInlineBotResultParams{QueryID: results.QueryID, Peer: chat, RandomID: results.QueryID, ID: res.ID})
	if err != nil {
		log.Println("Error sending inline result:", err)
		eOR(m, "<code>Coudn't fetch help menu</code>")
		return err
	}
	return err
}

func HelpCbk(cb *telegram.InlineCallbackQuery) error {
	data := string(cb.Data)
	if !utils.IsIn64Array(sudoers, cb.Sender.ID) && cb.Sender.ID != ubId {
		_, err := cb.Client.AnswerCallbackQuery(cb.QueryID, "You are not allowed to use this bot", &telegram.CallbackOptions{Alert: true})
		return err
	}
	if strings.Contains(data, "help:") {
		parts := strings.Split(data, ":")
		module := strings.ReplaceAll(parts[1], "_", " ")
		handlers, exists := HelpMap[module]
		if !exists {
			log.Println("Error: Module not found in HelpMap for module:", module)
			return fmt.Errorf("module not found in HelpMap")
		}
		msg := fmt.Sprintf("Here are the commands for <b>%s</b>:\n\n", module)
		for _, h := range handlers {
			msg += fmt.Sprintf("<code>.%s</code> - %s\n", h.Command, h.Description)
		}

		pageIndex := parts[2]
		replyMarkup := telegram.NewKeyboard().NewRow(1,
			telegram.ButtonBuilder{}.Data("⬅ Back", fmt.Sprintf("help_page_%s", pageIndex)),
		).Build()

		_, err := cb.Edit(msg, &telegram.SendOptions{ReplyMarkup: replyMarkup, ParseMode: "html"})
		return err
	}
	if strings.Contains(data, "help_page_") {
		parts := strings.Split(data, "_")
		index, _ := strconv.Atoi(parts[2])
		_, err := cb.Edit("<b>Aᴠᴀɪʟᴀʙʟᴇ Hᴇʟᴘ Mᴏᴅᴜʟᴇs:</b>", &telegram.SendOptions{ReplyMarkup: PaginateHelp(index), ParseMode: "html"})
		return err
	}

	return nil
}

func LoadHelpHandler(c *telegram.Client) {
	LoadModulesOrder()
	tgbot.AddInlineCallbackHandler("help", HelpCbk)
	tgbot.On("inline:help", HelpInline)
	AddHandler(&Handler{Command: "help", Func: HelpCmd}, c)
}
