package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"sort"
	"strconv"
	"strings"

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

func fuzzyMatch(query, target string) int {
	query = strings.ToLower(query)
	target = strings.ToLower(target)

	if query == target {
		return 100
	}
	if strings.HasPrefix(target, query) {
		return 90
	}
	if strings.Contains(target, query) {
		return 70
	}

	queryRunes := []rune(query)
	targetRunes := []rune(target)
	matchCount := 0
	targetIdx := 0

	for _, qr := range queryRunes {
		for targetIdx < len(targetRunes) {
			if targetRunes[targetIdx] == qr {
				matchCount++
				targetIdx++
				break
			}
			targetIdx++
		}
	}

	if matchCount == len(queryRunes) {
		return 50 + (matchCount * 10 / len(targetRunes))
	}

	return 0
}

func findBestMatch(query string) (string, []Handler, bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", nil, false
	}

	bestScore := 0
	var bestModule string
	var bestHandlers []Handler

	for module, handlers := range HelpMap {
		score := fuzzyMatch(query, module)
		if score > bestScore {
			bestScore = score
			bestModule = module
			bestHandlers = handlers
		}
	}

	for module, handlers := range HelpMap {
		for _, h := range handlers {
			score := fuzzyMatch(query, h.Command)
			if score > bestScore {
				bestScore = score
				bestModule = module
				bestHandlers = handlers
			}
		}
	}

	if bestScore >= 50 {
		return bestModule, bestHandlers, true
	}

	return "", nil, false
}

func formatModuleHelp(module string, handlers []Handler) string {
	msg := strings.Replace(locales.Tr("help.commands_header"), "%s", module, 1)
	for _, h := range handlers {
		entry := locales.Tr("help.command_entry")
		entry = strings.Replace(entry, "%s", h.Command, 1)
		entry = strings.Replace(entry, "%s", h.Description, 1)
		msg += entry
	}
	return msg
}

func HelpInline(i *telegram.InlineQuery) error {
	b := i.Builder()

	if !utils.IsIn64Array(sudoers, i.Sender.ID) && i.Sender.ID != ubId {
		btn := telegram.ButtonBuilder{}
		ownerLink := "tg://user?id=" + strconv.FormatInt(ubId, 10)
		if client != nil && client.Me() != nil && client.Me().Username != "" {
			ownerLink = "t.me/" + client.Me().Username
		}
		b.Article(locales.Tr("help.not_allowed_title"), locales.Tr("help.not_allowed_desc"), locales.Tr("help.not_allowed_desc"),
			&telegram.ArticleOptions{ReplyMarkup: telegram.NewKeyboard().NewRow(1, btn.URL(locales.Tr("help.owner_btn"), ownerLink)).Build()})
		i.Answer(b.Results())
		return nil
	}

	query := strings.TrimPrefix(i.Query, "help")
	query = strings.TrimSpace(query)

	if query != "" {
		module := strings.ReplaceAll(query, "_", " ")

		handlers, exists := HelpMap[module]
		if !exists {
			b.Article(locales.Tr("help.not_allowed_title"), locales.Tr("help.module_not_found"), locales.Tr("help.module_not_found"), nil)
			i.Answer(b.Results())
			return nil
		}

		msg := formatModuleHelp(module, handlers)
		replyMarkup := telegram.NewKeyboard().NewRow(1,
			telegram.ButtonBuilder{}.Data(locales.Tr("help.back_btn"), "help_page_0"),
		).Build()

		b.Article(module+" Help", "Help for "+module, msg, &telegram.ArticleOptions{ReplyMarkup: replyMarkup, ID: "help_" + strings.ReplaceAll(module, " ", "_")})
		i.Answer(b.Results())
		return nil
	}

	b.Article("Help Menu", "Available Help Menu", locales.Tr("help.menu_title"), &telegram.ArticleOptions{ReplyMarkup: PaginateHelp(0), ID: "help"})
	i.Answer(b.Results())
	return nil
}

func PaginateHelp(index int) *telegram.ReplyInlineMarkup {
	b := telegram.ButtonBuilder{}
	var btns []telegram.KeyboardButton
	max := 6
	total := len(ModuleList)
	start := index * max
	end := min(start+max, total)

	for _, mod := range ModuleList[start:end] {
		btns = append(btns, b.Data(mod, "help:"+strings.ReplaceAll(mod, " ", "_")+":"+strconv.Itoa(index)))
	}

	if index > 0 {
		btns = append(btns, b.Data(locales.Tr("help.back_btn"), "help_page_"+strconv.Itoa(index-1)))
	}
	if end < total {
		btns = append(btns, b.Data(locales.Tr("help.next_btn"), "help_page_"+strconv.Itoa(index+1)))
	}

	return telegram.NewKeyboard().NewGrid(4, 2, btns...).Build()
}

func HelpCmd(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())

	if args != "" {
		module, _, found := findBestMatch(args)
		if found {
			inlineQuery := "help " + strings.ReplaceAll(module, " ", "_")
			results, err := m.Client.InlineQuery(tbotId, &telegram.InlineOptions{Query: inlineQuery})
			if err != nil || len(results.Results) == 0 {
				handlers := HelpMap[module]
				msg := formatModuleHelp(module, handlers)
				_, err := eOR(m, msg)
				return err
			}

			res := results.Results[0].(*telegram.BotInlineResultObj)
			defer m.Delete()

			chat, _ := m.Client.GetSendablePeer(m.ChatID())
			_, err = m.Client.MessagesSendInlineBotResult(&telegram.MessagesSendInlineBotResultParams{
				QueryID: results.QueryID, Peer: chat, RandomID: results.QueryID, ID: res.ID,
			})
			if err != nil {
				logger.Error("Help module error:", err)
				handlers := HelpMap[module]
				msg := formatModuleHelp(module, handlers)
				_, _ = eOR(m, msg)
			}
			return err
		}

		_, err := eOR(m, locales.Tr("help.module_not_found"))
		return err
	}

	results, err := m.Client.InlineQuery(tbotId, &telegram.InlineOptions{Query: "help"})
	if err != nil || len(results.Results) == 0 {
		text := locales.Tr("help.menu_title") + "\n\n"
		for _, mod := range ModuleList {
			text += "• <b>" + mod + "</b>\n"
		}
		text += "\n" + locales.Tr("help.usage_hint")
		_, err := eOR(m, text)
		return err
	}

	res := results.Results[0].(*telegram.BotInlineResultObj)
	defer m.Delete()

	chat, _ := m.Client.GetSendablePeer(m.ChatID())
	_, err = m.Client.MessagesSendInlineBotResult(&telegram.MessagesSendInlineBotResultParams{
		QueryID: results.QueryID, Peer: chat, RandomID: results.QueryID, ID: res.ID,
	})
	if err != nil {
		logger.Error("Help error:", err)
		eOR(m, locales.Tr("help.fetch_error"))
	}
	return err
}

func HelpCbk(cb *telegram.InlineCallbackQuery) error {
	data := string(cb.Data)

	if !utils.IsIn64Array(sudoers, cb.Sender.ID) && cb.Sender.ID != ubId {
		cb.Client.AnswerCallbackQuery(cb.QueryID, locales.Tr("help.not_allowed_desc"), &telegram.CallbackOptions{Alert: true})
		return nil
	}

	if strings.Contains(data, "help:") {
		parts := strings.Split(data, ":")
		module := strings.ReplaceAll(parts[1], "_", " ")
		handlers, exists := HelpMap[module]
		if !exists {
			return nil
		}

		msg := formatModuleHelp(module, handlers)

		replyMarkup := telegram.NewKeyboard().NewRow(1,
			telegram.ButtonBuilder{}.Data(locales.Tr("help.back_btn"), "help_page_"+parts[2]),
		).Build()

		cb.Edit(msg, &telegram.SendOptions{ReplyMarkup: replyMarkup, ParseMode: "html"})
		return nil
	}

	if strings.Contains(data, "help_page_") {
		parts := strings.Split(data, "_")
		index, _ := strconv.Atoi(parts[2])
		cb.Edit(locales.Tr("help.menu_title"), &telegram.SendOptions{ReplyMarkup: PaginateHelp(index), ParseMode: "html"})
	}

	return nil
}

func LoadHelpHandler(c *telegram.Client) {
	LoadModulesOrder()
	tgbot.AddInlineCallbackHandler("help", HelpCbk)
	tgbot.On("inline:help", HelpInline)
	AddHandler(&Handler{Command: "help", Func: HelpCmd, Description: "Show help menu or search modules/commands", ModuleName: "Core"}, c)
}
