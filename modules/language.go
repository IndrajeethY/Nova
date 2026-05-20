package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Language", loadLanguageModule)
}

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
		_, err := eOR(m, locales.Trf("lang_settings.not_found", lang))
		return err
	}

	if err := locales.GetInstance().SetGlobalLanguage(lang); err != nil {
		_, err = eOR(m, locales.Tr("lang_settings.set_error"))
		return err
	}

	langName := locales.GetLanguageName(lang)
	_, err := eOR(m, locales.Trf("lang_settings.changed", langName, lang))
	return err
}

func GetLanguage(m *telegram.NewMessage) error {
	langs := locales.GetAvailableLanguages()
	defaultLang := "en"
	for _, l := range langs {
		if locales.Tr("language.code") == l {
			defaultLang = l
			break
		}
	}
	langName := locales.GetLanguageName(defaultLang)
	_, err := eOR(m, locales.Trf("lang_settings.current", langName, defaultLang))
	return err
}

func loadLanguageModule() {
	handlers := []*Handler{
		{Func: SetLanguage, Command: "setlang", Description: locales.Tr("desc.setlang"), ModuleName: "Language"},
		{Func: GetLanguage, Command: "lang", Description: locales.Tr("desc.lang"), ModuleName: "Language"},
	}
	AddHandlers(handlers, client)
}
