package modules

import (
	"NovaUserbot/locales"
	"NovaUserbot/utils"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

var timeFormatRegex = regexp.MustCompile(`^([0-5]?\d:)?([0-5]?\d:)?[0-5]?\d(\.\d+)?$`)

func checkFFmpeg() bool {
	_, err := utils.RunCommand("ffmpeg -version")
	return err == nil
}

func isValidTimeFormat(t string) bool {
	if !timeFormatRegex.MatchString(t) {
		return false
	}
	parts := strings.Split(strings.Split(t, ".")[0], ":")
	for _, p := range parts {
		if len(p) > 0 {
			val := 0
			for _, c := range p {
				if c < '0' || c > '9' {
					return false
				}
				val = val*10 + int(c-'0')
			}
			if val > 59 && len(parts) > 1 {
				return false
			}
		}
	}
	return true
}

func audioTrimCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("audiotools.trim_usage"))
		return err
	}

	parts := strings.Fields(args)
	if len(parts) < 2 {
		_, err := eOR(m, locales.Tr("audiotools.trim_usage"))
		return err
	}

	startTime := parts[0]
	endTime := parts[1]

	if !isValidTimeFormat(startTime) || !isValidTimeFormat(endTime) {
		_, err := eOR(m, locales.Tr("audiotools.invalid_time_format"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Audio() == nil && reply.Voice() == nil && reply.Video() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_audio"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.processing"))

	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".mp3"
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("trim_%d%s", m.ID, ext))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -ss %s -to %s -c copy %q", inputPath, startTime, endTime, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func audioConvertCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	targetFormat := strings.ToLower(strings.TrimSpace(m.Args()))
	validFormats := map[string]bool{"mp3": true, "wav": true, "ogg": true, "flac": true, "aac": true, "m4a": true, "opus": true}
	if targetFormat == "" || !validFormats[targetFormat] {
		_, err := eOR(m, locales.Tr("audiotools.convert_usage"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Audio() == nil && reply.Voice() == nil && reply.Video() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_audio"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.converting"))

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("convert_%d.%s", m.ID, targetFormat))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q %q", inputPath, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func extractAudioCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Video() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_video"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.extracting"))

	format := strings.ToLower(strings.TrimSpace(m.Args()))
	if format == "" {
		format = "mp3"
	}
	validFormats := map[string]bool{"mp3": true, "wav": true, "ogg": true, "flac": true, "aac": true}
	if !validFormats[format] {
		format = "mp3"
	}

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("extract_%d.%s", m.ID, format))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -vn -acodec libmp3lame -q:a 2 %q", inputPath, outputPath)
	if format != "mp3" {
		cmd = fmt.Sprintf("ffmpeg -y -i %q -vn %q", inputPath, outputPath)
	}

	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func audioBitrateCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	bitrate := strings.TrimSpace(m.Args())
	if bitrate == "" {
		bitrate = "128"
	}

	bitrateNum, err := strconv.Atoi(strings.TrimSuffix(bitrate, "k"))
	if err != nil || bitrateNum < 32 || bitrateNum > 320 {
		_, err := eOR(m, locales.Tr("audiotools.bitrate_usage"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Audio() == nil && reply.Voice() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_audio"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.processing"))

	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".mp3"
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("bitrate_%d%s", m.ID, ext))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -b:a %dk %q", inputPath, bitrateNum, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func voiceToMp3Command(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Voice() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_voice"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.converting"))

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("voice_%d.mp3", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -acodec libmp3lame -q:a 2 %q", inputPath, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func mp3ToVoiceCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Audio() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_audio"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.converting"))

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("tovoice_%d.ogg", m.ID))
	defer os.Remove(outputPath)

	cmd := fmt.Sprintf("ffmpeg -y -i %q -acodec libopus -b:a 64k %q", inputPath, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	uploaded, err := m.Client.UploadFile(outputPath)
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	_, err = m.Client.SendMedia(m.ChatID(), &telegram.InputMediaUploadedDocument{
		File:     uploaded,
		MimeType: "audio/ogg",
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeAudio{
				Voice: true,
			},
		},
	})

	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func audioSpeedCommand(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("audiotools.reply_required"))
		return err
	}

	speedStr := strings.TrimSpace(m.Args())
	if speedStr == "" {
		speedStr = "1.5"
	}

	speed, err := strconv.ParseFloat(speedStr, 64)
	if err != nil || speed < 0.5 || speed > 2.0 {
		_, err := eOR(m, locales.Tr("audiotools.speed_usage"))
		return err
	}

	if !checkFFmpeg() {
		_, err := eOR(m, locales.Tr("audiotools.ffmpeg_missing"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("audiotools.fetch_error"))
		return err
	}

	if reply.Audio() == nil && reply.Voice() == nil {
		_, err := eOR(m, locales.Tr("audiotools.no_audio"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("audiotools.downloading"))

	inputPath, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.download_error"))
		return err
	}
	defer os.Remove(inputPath)

	msg.Edit(locales.Tr("audiotools.processing"))

	ext := filepath.Ext(inputPath)
	if ext == "" {
		ext = ".mp3"
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("speed_%d%s", m.ID, ext))
	defer os.Remove(outputPath)

	atempo := speed
	cmd := fmt.Sprintf("ffmpeg -y -i %q -filter:a 'atempo=%f' %q", inputPath, atempo, outputPath)
	output, err := utils.RunCommand(cmd)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("audiotools.process_error"), output))
		return err
	}

	msg.Edit(locales.Tr("audiotools.uploading"))

	_, err = m.Respond("", &telegram.SendOptions{Media: outputPath})
	if err != nil {
		_, err := msg.Edit(locales.Tr("audiotools.upload_error"))
		return err
	}

	msg.Delete()
	return nil
}

func LoadAudioToolsModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "atrim", Func: audioTrimCommand, Description: "Trim audio (start end in HH:MM:SS)", ModuleName: "AudioTools"},
		{Command: "aconvert", Func: audioConvertCommand, Description: "Convert audio format (mp3/wav/ogg/flac/aac)", ModuleName: "AudioTools"},
		{Command: "extract", Func: extractAudioCommand, Description: "Extract audio from video", ModuleName: "AudioTools"},
		{Command: "abitrate", Func: audioBitrateCommand, Description: "Change audio bitrate (32-320)", ModuleName: "AudioTools"},
		{Command: "vtomp3", Func: voiceToMp3Command, Description: "Convert voice message to MP3", ModuleName: "AudioTools"},
		{Command: "tovoice", Func: mp3ToVoiceCommand, Description: "Convert audio to voice message", ModuleName: "AudioTools"},
		{Command: "aspeed", Func: audioSpeedCommand, Description: "Change audio speed (0.5-2.0x)", ModuleName: "AudioTools"},
	}
	AddHandlers(handlers, c)
}
