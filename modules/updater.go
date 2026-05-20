package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

const (
	repoURL = "https://github.com/IndrajeethY/Nova.git"
	branch  = "dev"
)

func init() {
	RegisterModule("Updater", loadUpdaterModule)
}

func checkUpdate(m *telegram.NewMessage) error {
	msg, _ := eOR(m, locales.Tr("updater.checking"))

	_, err := utils.RunCommand(fmt.Sprintf("git fetch %s", repoURL))
	if err != nil {
		_, err = msg.Edit(locales.Trf("updater.update_error", err.Error()))
		return err
	}

	diffOutput, err := utils.RunCommand(fmt.Sprintf("git diff HEAD origin/%s", branch))
	if err != nil {
		_, err = msg.Edit(locales.Trf("updater.update_error", err.Error()))
		return err
	}

	if strings.TrimSpace(diffOutput) == "" {
		_, err = msg.Edit(locales.Tr("updater.up_to_date"))
		return err
	}

	args := m.Args()
	if args == "force" {
		_, _ = msg.Edit(locales.Tr("updater.updating"))
		if _, err := utils.RunCommand("git reset --hard"); err != nil {
			_, err = msg.Edit(locales.Trf("updater.update_error", err.Error()))
			return err
		}
		if _, err := utils.RunCommand(fmt.Sprintf("git pull %s %s", repoURL, branch)); err != nil {
			_, err = msg.Edit(locales.Trf("updater.update_error", err.Error()))
			return err
		}
		utils.RunCommand("go build -o main .")
		restartBot()
		return nil
	}

	_, err = msg.Edit(locales.Tr("updater.updates_available"))
	return err
}

func loadUpdaterModule() {
	AddHandler(&Handler{
		ModuleName:    "Updater",
		Command:       "update",
		Description:   locales.Tr("desc.update"),
		Func:          checkUpdate,
		DisAllowSudos: true,
	}, client)
}
