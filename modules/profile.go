package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func setNameCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("profile.setname_usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("profile.updating"))

	firstName := args
	lastName := ""

	if strings.Contains(args, "//") {
		parts := strings.SplitN(args, "//", 2)
		firstName = strings.TrimSpace(parts[0])
		lastName = strings.TrimSpace(parts[1])
	}

	_, err := m.Client.AccountUpdateProfile(firstName, lastName, "")

	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.update_error"), err.Error()))
		}
		return err
	}

	if msg != nil {
		msg.Edit(fmt.Sprintf(locales.Tr("profile.name_changed"), args))
	}
	return nil
}

func setBioCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("profile.setbio_usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("profile.updating"))

	_, err := m.Client.AccountUpdateProfile("", "", args)

	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.update_error"), err.Error()))
		}
		return err
	}

	if msg != nil {
		msg.Edit(fmt.Sprintf(locales.Tr("profile.bio_changed"), args))
	}
	return nil
}

func setPicCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("profile.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("profile.fetch_error"))
		return err
	}

	if reply.Photo() == nil && reply.Video() == nil && reply.Document() == nil {
		_, err := eOR(m, locales.Tr("profile.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("profile.downloading"))

	filePath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("profile.download_error"))
		}
		return err
	}
	defer os.Remove(filePath)

	if msg != nil {
		msg.Edit(locales.Tr("profile.uploading"))
	}

	file, err := m.Client.UploadFile(filePath)
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("profile.upload_error"))
		}
		return err
	}

	isVideo := reply.Video() != nil
	if !isVideo && reply.Document() != nil {
		doc := reply.Document()
		if doc.MimeType == "video/mp4" {
			isVideo = true
		}
	}

	var uploadErr error
	if isVideo {
		_, uploadErr = m.Client.PhotosUploadProfilePhoto(&telegram.PhotosUploadProfilePhotoParams{
			Video: file,
		})
	} else {
		_, uploadErr = m.Client.PhotosUploadProfilePhoto(&telegram.PhotosUploadProfilePhotoParams{
			File: file,
		})
	}

	if uploadErr != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.update_error"), uploadErr.Error()))
		}
		return uploadErr
	}

	if msg != nil {
		msg.Edit(locales.Tr("profile.pic_changed"))
	}
	return nil
}

func delPfpCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	count := int32(1)

	if args == "all" {
		count = 0
	} else if args != "" {
		if n, err := strconv.Atoi(args); err == nil && n > 0 {
			count = int32(n)
		}
	}

	msg, _ := eOR(m, locales.Tr("profile.deleting"))

	photos, err := m.Client.GetProfilePhotos(m.Sender.ID, &telegram.PhotosOptions{Limit: count})
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.fetch_error")+": %s", err.Error()))
		}
		return err
	}

	if len(photos) == 0 {
		if msg != nil {
			msg.Edit(locales.Tr("profile.no_pfp"))
		}
		return nil
	}

	var photoList []telegram.InputPhoto
	for _, userPhoto := range photos {
		if photoObj, ok := userPhoto.Photo.(*telegram.PhotoObj); ok {
			photoList = append(photoList, &telegram.InputPhotoObj{
				ID:            photoObj.ID,
				AccessHash:    photoObj.AccessHash,
				FileReference: photoObj.FileReference,
			})
		}
	}

	if len(photoList) == 0 {
		if msg != nil {
			msg.Edit(locales.Tr("profile.no_pfp"))
		}
		return nil
	}

	_, err = m.Client.PhotosDeletePhotos(photoList)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.update_error"), err.Error()))
		}
		return err
	}

	if msg != nil {
		msg.Edit(fmt.Sprintf(locales.Tr("profile.pfp_deleted"), len(photoList)))
	}
	return nil
}

func getPfpCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	var targetPeer any
	limit := int32(1)

	if m.IsReply() {
		reply, err := m.GetReplyMessage()
		if err != nil {
			_, err := eOR(m, locales.Tr("profile.fetch_error"))
			return err
		}
		if reply.Sender == nil {
			_, err := eOR(m, locales.Tr("profile.user_not_found"))
			return err
		}
		targetPeer = reply.Sender.ID
	} else if args != "" {
		parts := strings.Fields(args)

		if id, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			targetPeer = id
		} else {
			targetPeer = strings.TrimPrefix(parts[0], "@")
		}
		if len(parts) > 1 {
			if parts[1] == "all" {
				limit = 0
			} else if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				limit = int32(n)
			}
		}
	} else {

		if m.IsPrivate() {
			targetPeer = m.Sender.ID
		} else {
			targetPeer = m.ChatID()
		}
	}

	msg, _ := eOR(m, locales.Tr("profile.fetching"))

	photos, err := m.Client.GetProfilePhotos(targetPeer, &telegram.PhotosOptions{Limit: limit})
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("profile.fetch_error")+": %v", err))
		}
		return err
	}

	if len(photos) == 0 {
		if msg != nil {
			msg.Edit(locales.Tr("profile.no_pfp_found"))
		}
		return nil
	}

	if msg != nil {
		msg.Delete()
	}

	for _, userPhoto := range photos {
		filePath, err := m.Client.DownloadMedia(&telegram.MessageMediaPhoto{Photo: userPhoto.Photo})
		if err != nil {
			continue
		}
		m.Respond("", &telegram.SendOptions{Media: filePath})
		os.Remove(filePath)
	}

	return nil
}

func LoadProfileModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "setname", Func: setNameCommand, Description: "Change profile name (first / last name)", ModuleName: "Profile", DisAllowSudos: true},
		{Command: "setbio", Func: setBioCommand, Description: "Change profile bio", ModuleName: "Profile", DisAllowSudos: true},
		{Command: "setpic", Func: setPicCommand, Description: "Change profile picture (reply to media)", ModuleName: "Profile", DisAllowSudos: true},
		{Command: "delpfp", Func: delPfpCommand, Description: "Delete profile picture(s) (n/all)", ModuleName: "Profile", DisAllowSudos: true},
		{Command: "poto", Func: getPfpCommand, Description: "Get profile picture(s)", ModuleName: "Profile"},
	}
	AddHandlers(handlers, c)
}
