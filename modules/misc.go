package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"os"

	"github.com/amarnathcjd/gogram/telegram"
)

func init() {
	RegisterModule("Misc", loadMiscModule)
}

func GenLink(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("misc.reply_to_media"))
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("misc.error_fetching_reply"))
		return err
	}
	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("misc.no_media"))
		return err
	}
	msg, _ := eOR(m, locales.Tr("misc.downloading"))
	file, err := reply.Download()
	if err != nil {
		_, err := eOR(m, locales.Tr("misc.error_downloading"))
		return err
	}
	defer os.Remove(file)
	msg.Edit(locales.Tr("misc.uploading"))
	link, err := utils.UploadFileToEnvsSh(file)
	if err != nil {
		_, err := eOR(m, err.Error())
		return err
	}
	_, err = msg.Edit(locales.Trf("misc.upload_success", link))
	return err
}

func loadMiscModule() {
	AddHandler(&Handler{
		Command:     "genlink",
		Description: "Generate a link for replied media",
		Func:        GenLink,
		ModuleName:  "Misc",
	}, client)
}
