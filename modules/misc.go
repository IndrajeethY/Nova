package modules

import (
	"NovaUserbot/utils"
	"fmt"
	"os"

	"github.com/amarnathcjd/gogram/telegram"
)

func GenLink(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, "Reply to a media message to generate a link.")
		return err
	}
	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, "An error occurred while fetching the reply message.")
		return err
	}
	if reply.Media() == nil {
		_, err := eOR(m, "The replied message does not contain any media.")
		return err
	}
	msg, _ := eOR(m, "Downloading media...")
	file, err := reply.Download()
	defer os.Remove(file)
	if err != nil {
		_, err := eOR(m, "An error occurred while downloading the media.")
		return err
	}
	msg.Edit("Uploading media...")
	link, err := utils.UploadFileToEnvsSh(file)
	if err != nil {
		_, err := eOR(m, err.Error())
		return err
	}
	_, err = msg.Edit(fmt.Sprintf("Successfully uploaded media.\nLink: %s", link))
	return err

}

func LoadMisc(c *telegram.Client) {
	gen := &Handler{
		Command:     "genlink",
		Description: "Generate a link for the replied media.",
		Func:        GenLink,
		ModuleName:  "Misc",
	}
	AddHandler(gen, c)
}
