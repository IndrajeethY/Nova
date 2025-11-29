package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func SetLanguage(m *telegram.NewMessage) error {
	lang := strings.TrimSpace(strings.ToLower(m.Args()))

	if lang == "" {
		langs := locales.GetAvailableLanguages()
		msg := locales.Tr("lang_settings.available_header")
		for _, l := range langs {
			name := locales.GetLanguageName(l)
			msg += fmt.Sprintf(locales.Tr("lang_settings.available_entry"), l, name)
		}
		msg += locales.Tr("lang_settings.usage")
		_, err := eOR(m, msg)
		return err
	}

	available := locales.GetAvailableLanguages()
	found := false
	for _, l := range available {
		if l == lang {
			found = true
			break
		}
	}

	if !found {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("lang_settings.not_found"), lang))
		return err
	}

	if err := db.Set("BOT_LANGUAGE", lang); err != nil {
		_, err = eOR(m, locales.Tr("lang_settings.set_error"))
		return err
	}

	locales.GetInstance().SetGlobalLanguage(lang)
	langName := locales.GetLanguageName(lang)
	_, err := eOR(m, fmt.Sprintf(locales.Tr("lang_settings.changed"), langName, lang))
	return err
}

func GetLanguage(m *telegram.NewMessage) error {
	lang := db.Get("BOT_LANGUAGE")
	if lang == "" {
		lang = "en"
	}
	langName := locales.GetLanguageName(lang)
	_, err := eOR(m, fmt.Sprintf(locales.Tr("lang_settings.current"), langName, lang))
	return err
}

func LoadLanguageModule(c *telegram.Client) {
	handlers := []*Handler{
		{Func: SetLanguage, Command: "setlang", Description: "Set bot language", ModuleName: "Language"},
		{Func: GetLanguage, Command: "lang", Description: "Show current language", ModuleName: "Language"},
	}
	AddHandlers(handlers, c)
}
