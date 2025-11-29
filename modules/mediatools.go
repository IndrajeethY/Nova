package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func mediaInfoCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("mediatools.no_media"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("mediatools.analyzing"))

	filePath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(filePath)

	output, err := utils.RunCommand(fmt.Sprintf("mediainfo %q", filePath))
	if err != nil {

		output, err = utils.RunCommand(fmt.Sprintf("ffprobe -hide_banner %q 2>&1", filePath))
		if err != nil {

			output = getBasicFileInfo(reply)
		}
	}

	if len(output) > 4000 {
		output = output[:4000] + "\n...[truncated]"
	}

	result := fmt.Sprintf(locales.Tr("mediatools.info_result"), output)
	if msg != nil {
		_, err = msg.Edit(result, &telegram.SendOptions{ParseMode: "HTML"})
	}
	return err
}

func getBasicFileInfo(m *telegram.NewMessage) string {
	var sb strings.Builder
	sb.WriteString("📄 <b>Media Information</b>\n")
	sb.WriteString("────────────────────\n")

	if m.File != nil {
		if m.File.Name != "" {
			sb.WriteString(fmt.Sprintf("<b>Name:</b> <code>%s</code>\n", m.File.Name))
		}
		sb.WriteString(fmt.Sprintf("<b>Size:</b> <code>%s</code>\n", humanBytes(uint64(m.File.Size))))
		if m.File.FileID != "" {
			sb.WriteString(fmt.Sprintf("<b>FileID:</b> <code>%s</code>\n", m.File.FileID))
		}
	}

	switch media := m.Message.Media.(type) {
	case *telegram.MessageMediaPhoto:
		sb.WriteString("<b>Type:</b> <code>Photo</code>\n")
	case *telegram.MessageMediaDocument:
		if doc, ok := media.Document.(*telegram.DocumentObj); ok {
			sb.WriteString(fmt.Sprintf("<b>MIME:</b> <code>%s</code>\n", doc.MimeType))
			for _, attr := range doc.Attributes {
				switch a := attr.(type) {
				case *telegram.DocumentAttributeVideo:
					sb.WriteString("<b>Type:</b> <code>Video</code>\n")
					sb.WriteString(fmt.Sprintf("<b>Duration:</b> <code>%.0f seconds</code>\n", a.Duration))
					sb.WriteString(fmt.Sprintf("<b>Resolution:</b> <code>%dx%d</code>\n", a.W, a.H))
				case *telegram.DocumentAttributeAudio:
					sb.WriteString("<b>Type:</b> <code>Audio</code>\n")
					sb.WriteString(fmt.Sprintf("<b>Duration:</b> <code>%d seconds</code>\n", a.Duration))
					if a.Title != "" {
						sb.WriteString(fmt.Sprintf("<b>Title:</b> <code>%s</code>\n", a.Title))
					}
					if a.Performer != "" {
						sb.WriteString(fmt.Sprintf("<b>Performer:</b> <code>%s</code>\n", a.Performer))
					}
				case *telegram.DocumentAttributeAnimated:
					sb.WriteString("<b>Type:</b> <code>Animation</code>\n")
				case *telegram.DocumentAttributeSticker:
					sb.WriteString("<b>Type:</b> <code>Sticker</code>\n")
				}
			}
		}
	}

	return sb.String()
}

func videoRotateCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	args := strings.TrimSpace(m.Args())
	angle := 90
	if args != "" {
		if a, err := strconv.Atoi(args); err == nil {

			if a == 90 || a == 180 || a == 270 {
				angle = a
			}
		}
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("mediatools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	if reply.Video() == nil && reply.Photo() == nil {
		_, err := eOR(m, locales.Tr("mediatools.no_video_photo"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("mediatools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.processing"))
	}

	ext := filepath.Ext(inputPath)
	if ext == "" {
		if reply.Video() != nil {
			ext = ".mp4"
		} else {
			ext = ".png"
		}
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("rotate_%d%s", m.ID, ext))
	defer os.Remove(outputPath)

	var cmd string
	if reply.Video() != nil {

		transposeValue := "1"
		switch angle {
		case 90:
			transposeValue = "1"
		case 180:
			transposeValue = "1,transpose=1"
		case 270:
			transposeValue = "2"
		}
		cmd = fmt.Sprintf("ffmpeg -y -i %q -vf 'transpose=%s' %q", inputPath, transposeValue, outputPath)
	} else {

		if !checkImageMagick() {
			if msg != nil {
				msg.Edit(locales.Tr("imagetools.imagemagick_missing"))
			}
			return nil
		}
		cmd = fmt.Sprintf("convert %q -rotate %d %q", inputPath, angle, outputPath)
	}

	output, err := utils.RunCommand(cmd)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("mediatools.process_error"), output))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func videoCompressCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("mediatools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	if reply.Video() == nil {
		_, err := eOR(m, locales.Tr("mediatools.no_video"))
		return err
	}

	args := strings.TrimSpace(m.Args())
	crf := "28"
	if args != "" {
		if n, err := strconv.Atoi(args); err == nil && n >= 0 && n <= 51 {
			crf = args
		}
	}

	msg, _ := eOR(m, locales.Tr("mediatools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.compressing"))
	}

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("compress_%d.mp4", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -c:v libx264 -crf %s -preset medium -c:a aac %q", inputPath, crf, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("mediatools.process_error"), output))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func videoToGifCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("mediatools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	if reply.Video() == nil {
		_, err := eOR(m, locales.Tr("mediatools.no_video"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("mediatools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.converting"))
	}

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("gif_%d.gif", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -vf 'fps=15,scale=320:-1:flags=lanczos' -gifflags +transdiff %q", inputPath, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("mediatools.process_error"), output))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func gifToVideoCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("mediatools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("mediatools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.converting"))
	}

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("video_%d.mp4", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -movflags faststart -pix_fmt yuv420p -vf 'scale=trunc(iw/2)*2:trunc(ih/2)*2' %q", inputPath, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("mediatools.process_error"), output))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func videoTrimCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("mediatools.reply_required"))
		return err
	}

	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("mediatools.vtrim_usage"))
		return err
	}

	parts := strings.Fields(args)
	if len(parts) < 2 {
		_, err := eOR(m, locales.Tr("mediatools.vtrim_usage"))
		return err
	}

	startTime := parts[0]
	endTime := parts[1]

	if !isValidTimeFormat(startTime) || !isValidTimeFormat(endTime) {
		_, err := eOR(m, locales.Tr("audiotools.invalid_time_format"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("mediatools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("mediatools.fetch_error"))
		return err
	}

	if reply.Video() == nil {
		_, err := eOR(m, locales.Tr("mediatools.no_video"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("mediatools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.download_error"))
		}
		return err
	}
	defer os.Remove(inputPath)

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.processing"))
	}

	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".mp4"
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("vtrim_%d%s", m.ID, ext))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -ss %s -to %s -c copy %q", inputPath, startTime, endTime, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		if msg != nil {
			msg.Edit(fmt.Sprintf(locales.Tr("mediatools.process_error"), output))
		}
		return err
	}

	if msg != nil {
		msg.Edit(locales.Tr("mediatools.uploading"))
	}

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		if msg != nil {
			msg.Edit(locales.Tr("mediatools.upload_error"))
		}
		return err
	}

	if msg != nil {
		msg.Delete()
	}
	return nil
}

func LoadMediaToolsModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "mediainfo", Func: mediaInfoCommand, Description: "Get media file information", ModuleName: "MediaTools"},
		{Command: "vrotate", Func: videoRotateCommand, Description: "Rotate video/image (90/180/270)", ModuleName: "MediaTools"},
		{Command: "vcompress", Func: videoCompressCommand, Description: "Compress video (CRF 0-51)", ModuleName: "MediaTools"},
		{Command: "vtogif", Func: videoToGifCommand, Description: "Convert video to GIF", ModuleName: "MediaTools"},
		{Command: "giftov", Func: gifToVideoCommand, Description: "Convert GIF to video", ModuleName: "MediaTools"},
		{Command: "vtrim", Func: videoTrimCommand, Description: "Trim video (start end in HH:MM:SS)", ModuleName: "MediaTools"},
	}
	AddHandlers(handlers, c)
}
