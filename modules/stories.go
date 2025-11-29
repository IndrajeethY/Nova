package modules

import (
	"NovaUserbot/locales"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

const maxStoriesToDownload = 5

func setStory(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("stories.reply_required"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil {
		_, err := eOR(m, locales.Tr("stories.error_fetching_reply"))
		return err
	}

	if reply.Media() == nil {
		_, err := eOR(m, locales.Tr("stories.no_media"))
		return err
	}

	args := strings.ToLower(strings.TrimSpace(m.Args()))
	var privacyRules []telegram.InputPrivacyRule

	switch args {
	case "contacts":
		privacyRules = []telegram.InputPrivacyRule{&telegram.InputPrivacyValueAllowContacts{}}
	case "all", "":
		privacyRules = []telegram.InputPrivacyRule{&telegram.InputPrivacyValueAllowAll{}}
	default:
		_, err := eOR(m, locales.Tr("stories.privacy_usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("stories.uploading"))

	file, err := reply.Download()
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.upload_error"), err.Error()))
		return err
	}
	defer os.Remove(file)

	uploaded, err := m.Client.UploadFile(file)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.upload_error"), err.Error()))
		return err
	}

	var inputMedia telegram.InputMedia
	if reply.Video() != nil {
		inputMedia = &telegram.InputMediaUploadedDocument{
			File:     uploaded,
			MimeType: "video/mp4",
		}
	} else {
		inputMedia = &telegram.InputMediaUploadedPhoto{
			File: uploaded,
		}
	}

	_, err = m.Client.StoriesSendStory(&telegram.StoriesSendStoryParams{
		Peer:         &telegram.InputPeerSelf{},
		Media:        inputMedia,
		PrivacyRules: privacyRules,
		RandomID:     time.Now().UnixNano(),
	})

	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.upload_error"), err.Error()))
		return err
	}

	privacyText := "all"
	if args == "contacts" {
		privacyText = "contacts"
	}
	_, err = msg.Edit(fmt.Sprintf(locales.Tr("stories.story_live_with_privacy"), privacyText))
	return err
}

func downloadStory(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	var username string
	var storyID int

	storyLinkPattern := regexp.MustCompile(`https?://t\.me/([^/]+)/s/(\d+)`)

	if args != "" {
		if match := storyLinkPattern.FindStringSubmatch(args); match != nil {
			username = match[1]
			storyID, _ = strconv.Atoi(match[2])
		} else {
			username = strings.TrimPrefix(args, "@")
		}
	} else if m.IsReply() {
		reply, err := m.GetReplyMessage()
		if err == nil && reply.Sender != nil {
			if reply.Sender.Username != "" {
				username = reply.Sender.Username
			} else {
				username = strconv.FormatInt(reply.Sender.ID, 10)
			}
		}
	}

	if username == "" {
		_, err := eOR(m, locales.Tr("stories.usage"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("stories.fetching"))

	peer, err := m.Client.GetSendablePeer(username)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.user_not_found"), username))
		return err
	}

	var peerInput telegram.InputPeer
	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		peerInput = &telegram.InputPeerUser{UserID: p.UserID, AccessHash: p.AccessHash}
	case *telegram.InputPeerChannel:
		peerInput = &telegram.InputPeerChannel{ChannelID: p.ChannelID, AccessHash: p.AccessHash}
	default:
		_, err := msg.Edit(locales.Tr("stories.invalid_peer"))
		return err
	}

	if storyID > 0 {
		stories, err := m.Client.StoriesGetStoriesByID(peerInput, []int32{int32(storyID)})
		if err != nil {
			_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.fetch_error"), err.Error()))
			return err
		}

		if len(stories.Stories) == 0 {
			_, err := msg.Edit(locales.Tr("stories.story_not_found"))
			return err
		}

		for _, story := range stories.Stories {
			if storyItem, ok := story.(*telegram.StoryItemObj); ok {
				err := downloadAndSendStoryItem(m, storyItem)
				if err != nil {
					continue
				}
			}
		}

		_, err = msg.Edit(locales.Tr("stories.uploaded"))
		return err
	}

	var storyItems []telegram.StoryItem
	switch p := peer.(type) {
	case *telegram.InputPeerUser:
		fullUser, err := m.Client.UsersGetFullUser(&telegram.InputUserObj{UserID: p.UserID, AccessHash: p.AccessHash})
		if err != nil {
			_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.fetch_error"), err.Error()))
			return err
		}
		if fullUser.FullUser.Stories != nil {
			storyItems = fullUser.FullUser.Stories.Stories
		}
	case *telegram.InputPeerChannel:
		fullChannel, err := m.Client.ChannelsGetFullChannel(&telegram.InputChannelObj{ChannelID: p.ChannelID, AccessHash: p.AccessHash})
		if err != nil {
			_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.fetch_error"), err.Error()))
			return err
		}
		if cf, ok := fullChannel.FullChat.(*telegram.ChannelFull); ok && cf.Stories != nil {
			storyItems = cf.Stories.Stories
		}
	}

	if len(storyItems) == 0 {
		_, err := msg.Edit(locales.Tr("stories.no_stories"))
		return err
	}

	count := 0
	for _, story := range storyItems {
		if count >= maxStoriesToDownload {
			break
		}
		if storyItem, ok := story.(*telegram.StoryItemObj); ok {
			err := downloadAndSendStoryItem(m, storyItem)
			if err != nil {
				continue
			}
			count++
		}
	}

	if count == 0 {
		_, err := msg.Edit(locales.Tr("stories.download_failed"))
		return err
	}

	_, err = msg.Edit(fmt.Sprintf(locales.Tr("stories.uploaded_count"), count))
	return err
}

func downloadAndSendStoryItem(m *telegram.NewMessage, story *telegram.StoryItemObj) error {
	if story.Media == nil {
		return fmt.Errorf("no media in story")
	}

	ext := ".jpg"
	isVideo := false
	switch media := story.Media.(type) {
	case *telegram.MessageMediaDocument:
		if doc, ok := media.Document.(*telegram.DocumentObj); ok {
			if doc.MimeType == "video/mp4" || strings.HasPrefix(doc.MimeType, "video/") {
				ext = ".mp4"
				isVideo = true
			}
		}
	}

	tmpFile := fmt.Sprintf("/tmp/story_%d_%d%s", story.ID, time.Now().UnixNano(), ext)

	file, err := m.Client.DownloadMedia(story.Media, &telegram.DownloadOptions{
		FileName: tmpFile,
	})
	if err != nil {
		return err
	}
	defer os.Remove(file)

	caption := ""
	if story.Caption != "" {
		caption = story.Caption
	}

	opts := &telegram.SendOptions{Media: file}
	if isVideo {
		opts.Attributes = []telegram.DocumentAttribute{
			&telegram.DocumentAttributeVideo{
				SupportsStreaming: true,
			},
		}
	}

	_, err = m.Respond(caption, opts)

	return err
}

func downloadArchiveStory(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())

	index := 0
	if strings.Contains(args, "-n") {
		parts := strings.Split(args, "-n")
		if len(parts) > 1 {
			nVal := strings.TrimSpace(parts[1])
			nValParts := strings.Fields(nVal)
			if len(nValParts) > 0 {
				if n, err := strconv.Atoi(nValParts[0]); err == nil && n > 0 {
					index = n - 1
				}
			}
		}
	}

	msg, _ := eOR(m, locales.Tr("stories.fetching"))

	archived, err := m.Client.StoriesGetStoriesArchive(&telegram.InputPeerSelf{}, 0, 100)
	if err != nil {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.fetch_error"), err.Error()))
		return err
	}

	if len(archived.Stories) == 0 {
		_, err := msg.Edit(locales.Tr("stories.no_stories"))
		return err
	}

	if index >= len(archived.Stories) {
		_, err := msg.Edit(fmt.Sprintf(locales.Tr("stories.invalid_index"), len(archived.Stories), len(archived.Stories)))
		return err
	}

	story := archived.Stories[index]
	if storyItem, ok := story.(*telegram.StoryItemObj); ok {
		err := downloadAndSendStoryItem(m, storyItem)
		if err != nil {
			_, err := msg.Edit(locales.Tr("stories.download_failed"))
			return err
		}
		_, err = msg.Edit(fmt.Sprintf(locales.Tr("stories.archived_downloaded"), index+1))
		return err
	}

	_, err = msg.Edit(locales.Tr("stories.download_failed"))
	return err
}

func LoadStoriesModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "setstory", Description: "Set replied media as story (all/contacts)", Func: setStory, ModuleName: "Stories", DisAllowSudos: true},
		{Command: "storydl", Description: "Download user stories", Func: downloadStory, ModuleName: "Stories"},
		{Command: "archdl", Description: "Download archived story (-n <index> to pick)", Func: downloadArchiveStory, ModuleName: "Stories", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)
}
