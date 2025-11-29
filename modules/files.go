package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	downloadCancels = make(map[int32]context.CancelFunc)
	cancelMutex     sync.RWMutex
)

func humanBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func sendFileByIDCommand(m *telegram.NewMessage) error {
	fileId := strings.TrimSpace(m.Args())
	if fileId == "" {
		_, err := eOR(m, locales.Tr("files.no_fileid"))
		return err
	}

	file, err := telegram.ResolveBotFileID(fileId)
	if err != nil {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("files.resolve_error"), err.Error()))
		return err
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: file})
	return err
}

func getFileIDCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("files.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("files.fetch_error"))
		return err
	}

	if reply.File == nil {
		_, err := eOR(m, locales.Tr("files.no_file"))
		return err
	}

	_, err = eOR(m, fmt.Sprintf(locales.Tr("files.fileid_result"), reply.File.FileID))
	return err
}

func uploadCommand(m *telegram.NewMessage) error {
	filename := strings.TrimSpace(m.Args())
	if filename == "" {
		_, err := eOR(m, locales.Tr("files.no_filename"))
		return err
	}

	spoiler := false
	forceDoc := false

	if strings.Contains(filename, "-s") {
		spoiler = true
		filename = strings.ReplaceAll(filename, "-s", "")
		filename = strings.TrimSpace(filename)
	}
	if strings.Contains(filename, "--doc") {
		forceDoc = true
		filename = strings.ReplaceAll(filename, "--doc", "")
		filename = strings.TrimSpace(filename)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		_, err := eOR(m, locales.Tr("files.file_not_found"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("files.uploading"))
	uploadStartTimestamp := time.Now()

	opts := &telegram.SendOptions{
		Media:         filename,
		Spoiler:       spoiler,
		ForceDocument: forceDoc,
	}

	if _, err := m.Respond("", opts); err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("files.upload_error"), err.Error()))
		return err
	}

	_, err := msg.Edit(fmt.Sprintf(locales.Tr("files.upload_success"), filename, time.Since(uploadStartTimestamp).String()))
	return err
}

func downloadCommand(m *telegram.NewMessage) error {
	if !m.IsReply() && m.Args() == "" {
		_, err := eOR(m, locales.Tr("files.reply_or_link"))
		return err
	}

	fn := strings.TrimSpace(m.Args())

	var reply *telegram.NewMessage
	var msg *telegram.NewMessage

	if m.IsReply() {
		r, err := m.GetReplyMessage()
		if err != nil {
			_, err := eOR(m, locales.Tr("files.fetch_error"))
			return err
		}
		reply = r
		msg, _ = eOR(m, locales.Tr("files.downloading"))
	} else {

		reg := regexp.MustCompile(`t\.me/(\w+)/(\d+)`)
		match := reg.FindStringSubmatch(m.Args())
		if len(match) != 3 || match[1] == "c" {

			reg = regexp.MustCompile(`t\.me/c/(\d+)/(\d+)`)
			match = reg.FindStringSubmatch(m.Args())
			if len(match) != 3 {
				_, err := eOR(m, locales.Tr("files.invalid_link"))
				return err
			}

			msgId, err := strconv.Atoi(match[2])
			if err != nil {
				_, err := eOR(m, locales.Tr("files.invalid_link"))
				return err
			}

			chatID, err := strconv.Atoi(match[1])
			if err != nil {
				_, err := eOR(m, locales.Tr("files.invalid_link"))
				return err
			}

			msgX, err := m.Client.GetMessageByID(chatID, int32(msgId))
			if err != nil {
				_, err := eOR(m, fmt.Sprintf(locales.Tr("files.fetch_error")+": %s", err.Error()))
				return err
			}
			reply = msgX
			if reply.File != nil {
				fn = reply.File.Name
			}
			msg, _ = eOR(m, fmt.Sprintf(locales.Tr("files.downloading_from"), "private", msgId))
		} else {
			username := match[1]
			msgId, err := strconv.Atoi(match[2])
			if err != nil {
				_, err := eOR(m, locales.Tr("files.invalid_link"))
				return err
			}

			msgX, err := m.Client.GetMessageByID(username, int32(msgId))
			if err != nil {
				_, err := eOR(m, fmt.Sprintf(locales.Tr("files.fetch_error")+": %s", err.Error()))
				return err
			}
			reply = msgX
			if reply.File != nil {
				fn = reply.File.Name
			}
			msg, _ = eOR(m, fmt.Sprintf(locales.Tr("files.downloading_from"), username, msgId))
		}
	}

	if reply.File == nil {
		_, err := msg.Edit(locales.Tr("files.no_file"))
		return err
	}

	uploadStartTimestamp := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelMutex.Lock()
	downloadCancels[msg.ID] = cancel
	cancelMutex.Unlock()

	defer func() {
		cancelMutex.Lock()
		delete(downloadCancels, msg.ID)
		cancelMutex.Unlock()
	}()

	opts := &telegram.DownloadOptions{
		Ctx: ctx,
	}
	if fn != "" {
		opts.FileName = fn
	}

	filePath, err := reply.Download(opts)
	if err != nil {
		if err == context.Canceled {
			msg.Edit(locales.Tr("files.download_cancelled"))
		} else {
			msg.Edit(fmt.Sprintf(locales.Tr("files.download_error"), err.Error()))
		}
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("files.download_success"), filePath, time.Since(uploadStartTimestamp).String()))
	return err
}

func cancelDownloadCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("files.reply_to_download"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("files.fetch_error"))
		return err
	}

	cancelMutex.RLock()
	cancel, exists := downloadCancels[reply.ID]
	cancelMutex.RUnlock()

	if !exists {
		_, err := eOR(m, locales.Tr("files.no_active_download"))
		return err
	}

	cancel()
	_, err = eOR(m, locales.Tr("files.cancelled"))
	return err
}

func fileInfoCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("files.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("files.fetch_error"))
		return err
	}

	type fileInfo struct {
		FileName   string
		Type       string
		Size       int64
		FileID     string
		Attributes map[string]string
	}

	fi := fileInfo{
		Attributes: make(map[string]string),
	}

	if reply.File != nil {
		fi.FileName = reply.File.Name
		fi.Size = reply.File.Size
		fi.FileID = reply.File.FileID
	}

	switch media := reply.Message.Media.(type) {
	case *telegram.MessageMediaDocument:
		fi.Type = "Document"
		if doc, ok := media.Document.(*telegram.DocumentObj); ok {
			for _, attr := range doc.Attributes {
				switch a := attr.(type) {
				case *telegram.DocumentAttributeVideo:
					fi.Type = "Video"
					fi.Attributes["Duration"] = fmt.Sprintf("%.0f seconds", a.Duration)
					fi.Attributes["Width"] = fmt.Sprintf("%d px", a.W)
					fi.Attributes["Height"] = fmt.Sprintf("%d px", a.H)
				case *telegram.DocumentAttributeAudio:
					fi.Type = "Audio"
					fi.Attributes["Duration"] = fmt.Sprintf("%d seconds", a.Duration)
					if a.Title != "" {
						fi.Attributes["Title"] = a.Title
					}
					if a.Performer != "" {
						fi.Attributes["Performer"] = a.Performer
					}
					fi.Attributes["Voice"] = strconv.FormatBool(a.Voice)
				case *telegram.DocumentAttributeAnimated:
					fi.Type = "Animation"
				case *telegram.DocumentAttributeSticker:
					fi.Type = "Sticker"
					if a.Alt != "" {
						fi.Attributes["Alt"] = a.Alt
					}
				}
			}
		}
	case *telegram.MessageMediaPhoto:
		fi.Type = "Photo"
	case *telegram.MessageMediaPoll:
		fi.Type = "Poll"
	case *telegram.MessageMediaGeo:
		fi.Type = "Geo"
		if geo, ok := media.Geo.(*telegram.GeoPointObj); ok {
			fi.Attributes["AccuracyRadius"] = fmt.Sprintf("%d meters", geo.AccuracyRadius)
			fi.Attributes["Latitude"] = fmt.Sprintf("%.6f", geo.Lat)
			fi.Attributes["Longitude"] = fmt.Sprintf("%.6f", geo.Long)
		}
	default:
		fi.Type = "Unknown"
	}

	var output strings.Builder
	output.WriteString("📄 <b>File Information</b>\n")
	output.WriteString("────────────────────\n")
	if fi.FileName != "" {
		output.WriteString(fmt.Sprintf("📛 <b>FileName</b>: <code>%s</code>\n", fi.FileName))
	}
	output.WriteString(fmt.Sprintf("📂 <b>Type</b>: <code>%s</code>\n", fi.Type))
	if fi.Size > 0 {
		output.WriteString(fmt.Sprintf("📦 <b>Size</b>: <code>%s</code>\n", humanBytes(uint64(fi.Size))))
	}
	if fi.FileID != "" {
		output.WriteString(fmt.Sprintf("🆔 <b>FileID</b>: <code>%s</code>\n", fi.FileID))
	}
	if len(fi.Attributes) > 0 {
		output.WriteString("⚙️ <b>Attributes</b>:\n")
		for k, v := range fi.Attributes {
			output.WriteString(fmt.Sprintf("   • <b>%s</b>: <code>%s</code>\n", k, v))
		}
	}

	_, err = eOR(m, output.String())
	return err
}

func genLinkCommand(m *telegram.NewMessage) error {
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

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("misc.upload_success"), link))
	return err
}

func LoadFilesModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "file", Func: sendFileByIDCommand, Description: "Send a file by its FileID", ModuleName: "Files"},
		{Command: "fid", Func: getFileIDCommand, Description: "Get FileID of replied media", ModuleName: "Files"},
		{Command: "ul", Func: uploadCommand, Description: "Upload a file [-s for spoiler, --doc for document]", ModuleName: "Files"},
		{Command: "dl", Func: downloadCommand, Description: "Download replied file or from link", ModuleName: "Files"},
		{Command: "cancel", Func: cancelDownloadCommand, Description: "Cancel an active download", ModuleName: "Files"},
		{Command: "finfo", Func: fileInfoCommand, Description: "Get file information", ModuleName: "Files"},
		{Command: "genlink", Func: genLinkCommand, Description: "Generate shareable link for media", ModuleName: "Files"},
	}
	AddHandlers(handlers, c)
}
