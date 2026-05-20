package modules

import (
	"NovaUserbot/locales"
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
	helpMu.RLock()
	defer helpMu.RUnlock()
	ModuleList = nil
	for mod := range HelpMap {
		ModuleList = append(ModuleList, mod)
	}
	sort.Strings(ModuleList)
}

func HelpInline(i *telegram.InlineQuery) error {
	b := i.Builder()
	if !IsSudoer(i.Sender.ID) && i.Sender.ID != ubId {
		btn := telegram.ButtonBuilder{}
		b.Article(locales.Tr("help.not_allowed_title"), locales.Tr("help.not_allowed_desc"), locales.Tr("help.not_allowed_desc"),
			&telegram.ArticleOptions{ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL("Owner", "t.me/tamilvip007")).Build()})
		i.Answer(b.Results())
		return nil
	}
	b.Article("Help Menu", "Available Help Menu", locales.Tr("help.menu_title"),
		&telegram.ArticleOptions{ReplyMarkup: PaginateHelp(0), ID: "help"})
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
	results, err := m.Client.InlineQuery(tbotId, &telegram.InlineOptions{Query: "help"})
	if err != nil || results == nil || len(results.Results) == 0 {
		eOR(m, locales.Tr("help.fetch_error"))
		return err
	}
	res, ok := results.Results[0].(*telegram.BotInlineResultObj)
	if !ok {
		eOR(m, locales.Tr("help.fetch_error"))
		return nil
	}
	defer m.Delete()
	chat, _ := m.Client.GetSendablePeer(m.ChatID())
	_, err = m.Client.MessagesSendInlineBotResult(&telegram.MessagesSendInlineBotResultParams{
		QueryID: results.QueryID, Peer: chat, RandomID: results.QueryID, ID: res.ID,
	})
	if err != nil {
		log.Println("Error sending inline result:", err)
		eOR(m, locales.Tr("help.fetch_error"))
		return err
	}
	return nil
}

func HelpCbk(cb *telegram.InlineCallbackQuery) error {
	data := string(cb.Data)
	if !IsSudoer(cb.Sender.ID) && cb.Sender.ID != ubId {
		cb.Client.AnswerCallbackQuery(cb.QueryID, locales.Tr("help.not_allowed_desc"), &telegram.CallbackOptions{Alert: true})
		return nil
	}
	if strings.Contains(data, "help:") {
		parts := strings.Split(data, ":")
		if len(parts) < 3 {
			return nil
		}
		module := strings.ReplaceAll(parts[1], "_", " ")
		handlers, exists := GetHelpModule(module)
		if !exists {
			return fmt.Errorf("module not found in HelpMap")
		}
		msg := fmt.Sprintf(locales.Tr("help.commands_header"), module) + "\n\n"
		for _, h := range handlers {
			msg += fmt.Sprintf(locales.Tr("help.command_entry"), h.Command, h.Description) + "\n"
		}
		pageIndex := parts[2]
		replyMarkup := telegram.NewKeyboard().NewRow(1,
			telegram.ButtonBuilder{}.Data("⬅ Back", fmt.Sprintf("help_page_%s", pageIndex)),
		).Build()
		cb.Edit(msg, &telegram.SendOptions{ReplyMarkup: replyMarkup, ParseMode: "html"})
		return nil
	}
	if strings.Contains(data, "help_page_") {
		parts := strings.Split(data, "_")
		if len(parts) < 3 {
			return nil
		}
		index, _ := strconv.Atoi(parts[2])
		cb.Edit(locales.Tr("help.menu_title"), &telegram.SendOptions{ReplyMarkup: PaginateHelp(index), ParseMode: "html"})
		return nil
	}
	return nil
}

func init() {
	RegisterModule("Help", loadHelpModule, 100)
}

func loadHelpModule() {
	LoadModulesOrder()
	tgbot.AddInlineCallbackHandler("help", HelpCbk)
	tgbot.On("inline:help", HelpInline)
	AddHandler(&Handler{Command: "help", Description: locales.Tr("desc.help"), Func: HelpCmd, ModuleName: "Help"}, client)
}
