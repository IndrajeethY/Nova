package modules

import (
	"NovaUserbot/locales"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/nfnt/resize"
)

const (
	stickerPackKey  = "STICKER_PACK"
	stickerPacksKey = "STICKER_PACKS"
	maxStickers     = 120
)

func init() {
	RegisterModule("Stickers", loadStickersModule)
}

func getPackShortName() string {
	name, _ := Db.Get(context.Background(), stickerPackKey).Result()
	return name
}

func setPackShortName(name string) {
	Db.Set(context.Background(), stickerPackKey, name, 0)
	addPackToList(name)
}

func addPackToList(name string) {
	Db.SAdd(context.Background(), stickerPacksKey, name)
}

func getAllPacks() []string {
	packs, _ := Db.SMembers(context.Background(), stickerPacksKey).Result()
	return packs
}

func docToInputDoc(doc *telegram.DocumentObj) *telegram.InputDocumentObj {
	return &telegram.InputDocumentObj{
		ID:            doc.ID,
		AccessHash:    doc.AccessHash,
		FileReference: doc.FileReference,
	}
}

func getStickerEmoji(doc *telegram.DocumentObj) string {
	for _, attr := range doc.Attributes {
		if s, ok := attr.(*telegram.DocumentAttributeSticker); ok {
			if s.Alt != "" {
				return s.Alt
			}
		}
	}
	return "🤔"
}

func isImageDoc(m *telegram.NewMessage) bool {
	doc := m.Document()
	return doc != nil && strings.HasPrefix(doc.MimeType, "image/")
}

func nextPackShortName() string {
	base := fmt.Sprintf("nova_%d", ubId)
	for i := 1; ; i++ {
		name := base
		if i > 1 {
			name = fmt.Sprintf("%s_%d", base, i)
		}
		_, err := client.MessagesGetStickerSet(&telegram.InputStickerSetShortName{ShortName: name}, 0)
		if err != nil {
			return name
		}
	}
}

func findAvailablePack() (string, error) {
	packName := getPackShortName()
	if packName == "" {
		return "", nil
	}

	setResult, err := client.MessagesGetStickerSet(&telegram.InputStickerSetShortName{ShortName: packName}, 0)
	if err != nil {
		return "", nil
	}

	setObj, ok := setResult.(*telegram.MessagesStickerSetObj)
	if !ok {
		return "", nil
	}

	if setObj.Set.Count < maxStickers {
		return packName, nil
	}

	return "", nil
}

func createPack(emoji string, inputDoc telegram.InputDocument) (*telegram.MessagesStickerSetObj, error) {
	user, err := client.GetSendableUser(ubId)
	if err != nil {
		return nil, err
	}

	me := client.Me()
	shortName := nextPackShortName()
	num := 1
	if strings.Contains(shortName, fmt.Sprintf("nova_%d_", ubId)) {
		parts := strings.Split(shortName, "_")
		if len(parts) > 2 {
			num, _ = strconv.Atoi(parts[len(parts)-1])
		}
	}

	title := fmt.Sprintf("%s's Nova Pack", me.FirstName)
	if num > 1 {
		title = fmt.Sprintf("%s's Nova Pack Vol.%d", me.FirstName, num)
	}

	result, err := client.StickersCreateStickerSet(&telegram.StickersCreateStickerSetParams{
		UserID:    user,
		Title:     title,
		ShortName: shortName,
		Stickers:  []*telegram.InputStickerSetItem{{Document: inputDoc, Emoji: emoji}},
	})
	if err != nil {
		return nil, err
	}

	sObj, ok := result.(*telegram.MessagesStickerSetObj)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}
	setPackShortName(sObj.Set.ShortName)
	return sObj, nil
}

func addStickerToPackOrCreate(inputDoc telegram.InputDocument, emoji string) (string, error) {
	packName, err := findAvailablePack()
	if err != nil {
		return "", err
	}

	if packName == "" {
		result, err := createPack(emoji, inputDoc)
		if err != nil {
			return "", err
		}
		return result.Set.ShortName, nil
	}

	_, addErr := client.StickersAddStickerToSet(
		&telegram.InputStickerSetShortName{ShortName: packName},
		&telegram.InputStickerSetItem{Document: inputDoc, Emoji: emoji},
	)
	if addErr != nil {
		if strings.Contains(addErr.Error(), "STICKERS_TOO_MUCH") || strings.Contains(addErr.Error(), "STICKERPACK_STICKERS_TOO_MUCH") {
			result, err := createPack(emoji, inputDoc)
			if err != nil {
				return "", err
			}
			return result.Set.ShortName, nil
		}
		return "", addErr
	}
	return packName, nil
}

func uploadPhotoAsSticker(reply *telegram.NewMessage, emoji string) (telegram.InputDocument, error) {
	filePath, err := client.DownloadMedia(reply.Media())
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	file.Close()
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	resized := resize.Thumbnail(512, 512, img, resize.Lanczos3)

	var buf bytes.Buffer
	if err := png.Encode(&buf, resized); err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "sticker_*.png")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	inputFile, err := client.UploadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	peer, err := client.ResolvePeer(ubId)
	if err != nil {
		return nil, err
	}

	media, err := client.MessagesUploadMedia("", peer, &telegram.InputMediaUploadedDocument{
		File:     inputFile,
		MimeType: "image/png",
		Attributes: []telegram.DocumentAttribute{
			&telegram.DocumentAttributeFilename{FileName: "sticker.png"},
			&telegram.DocumentAttributeSticker{
				Alt:        emoji,
				Stickerset: &telegram.InputStickerSetEmpty{},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("upload media failed: %w", err)
	}

	docMedia, ok := media.(*telegram.MessageMediaDocument)
	if !ok {
		return nil, fmt.Errorf("unexpected media type")
	}
	doc, ok := docMedia.Document.(*telegram.DocumentObj)
	if !ok {
		return nil, fmt.Errorf("unexpected document type")
	}

	return docToInputDoc(doc), nil
}

func kangSticker(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.kang_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("stickers.kanging"))

	emoji := strings.TrimSpace(m.Args())
	var inputDoc telegram.InputDocument

	sticker := reply.Sticker()
	if sticker != nil {
		if emoji == "" {
			emoji = getStickerEmoji(sticker)
		}
		inputDoc = docToInputDoc(sticker)
	} else if reply.Photo() != nil || isImageDoc(reply) {
		if emoji == "" {
			emoji = "🤔"
		}
		inputDoc, err = uploadPhotoAsSticker(reply, emoji)
		if err != nil {
			_, err = msg.Edit(locales.Trf("stickers.kang_error", err.Error()))
			return err
		}
	} else {
		_, err = msg.Edit(locales.Tr("stickers.not_sticker"))
		return err
	}

	packName, err := addStickerToPackOrCreate(inputDoc, emoji)
	if err != nil {
		_, err = msg.Edit(locales.Trf("stickers.kang_error", err.Error()))
		return err
	}
	_, err = msg.Edit(locales.Trf("stickers.kanged", emoji, packName))
	return err
}

func kangPack(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.pkang_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	sticker := reply.Sticker()
	if sticker == nil {
		_, err = eOR(m, locales.Tr("stickers.not_sticker"))
		return err
	}

	var stickerSetInput telegram.InputStickerSet
	for _, attr := range sticker.Attributes {
		if s, ok := attr.(*telegram.DocumentAttributeSticker); ok {
			stickerSetInput = s.Stickerset
			break
		}
	}

	if stickerSetInput == nil {
		_, err = eOR(m, locales.Tr("stickers.no_pack"))
		return err
	}
	if _, ok := stickerSetInput.(*telegram.InputStickerSetEmpty); ok {
		_, err = eOR(m, locales.Tr("stickers.no_pack"))
		return err
	}

	stickerSet, err := client.MessagesGetStickerSet(stickerSetInput, 0)
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.fetch_error", err.Error()))
		return err
	}

	setObj, ok := stickerSet.(*telegram.MessagesStickerSetObj)
	if !ok {
		_, err = eOR(m, locales.Tr("stickers.error"))
		return err
	}

	targetPack := strings.TrimSpace(m.Args())
	msg, _ := eOR(m, locales.Trf("stickers.pkanging", setObj.Set.Title, len(setObj.Documents)))

	added := 0
	failed := 0

	if targetPack != "" {
		setPackShortName(targetPack)
	}

	for _, doc := range setObj.Documents {
		docObj, ok := doc.(*telegram.DocumentObj)
		if !ok {
			failed++
			continue
		}

		emoji := getStickerEmoji(docObj)
		inputDoc := docToInputDoc(docObj)

		packName, err := addStickerToPackOrCreate(inputDoc, emoji)
		if err != nil {
			failed++
			continue
		}
		_ = packName
		added++
	}

	currentPack := getPackShortName()
	_, err = msg.Edit(locales.Trf("stickers.pkanged", added, failed, currentPack))
	return err
}

func listPacksCmd(m *telegram.NewMessage) error {
	packs := getAllPacks()
	if len(packs) == 0 {
		_, err := eOR(m, locales.Tr("stickers.no_packs"))
		return err
	}

	current := getPackShortName()
	text := locales.Tr("stickers.listpacks_header") + "\n\n"
	for i, pack := range packs {
		marker := ""
		if pack == current {
			marker = " ✅"
		}
		text += fmt.Sprintf(locales.Tr("stickers.listpacks_entry"), i+1, pack, pack, marker) + "\n"
	}

	_, err := eOR(m, text, &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func setStickerEmoji(m *telegram.NewMessage) error {
	emoji := strings.TrimSpace(m.Args())
	if emoji == "" || !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.setemoji_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	sticker := reply.Sticker()
	if sticker == nil {
		_, err = eOR(m, locales.Tr("stickers.not_sticker"))
		return err
	}

	_, err = client.StickersChangeSticker(docToInputDoc(sticker), emoji, nil, "")
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.setemoji_error", err.Error()))
		return err
	}
	_, err = eOR(m, locales.Trf("stickers.setemoji_success", emoji))
	return err
}

func renamePackCmd(m *telegram.NewMessage) error {
	title := strings.TrimSpace(m.Args())
	if title == "" {
		_, err := eOR(m, locales.Tr("stickers.packname_usage"))
		return err
	}

	packName := getPackShortName()
	if packName == "" {
		_, err := eOR(m, locales.Tr("stickers.no_user_pack"))
		return err
	}

	_, err := client.StickersRenameStickerSet(
		&telegram.InputStickerSetShortName{ShortName: packName},
		title,
	)
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.packname_error", err.Error()))
		return err
	}
	_, err = eOR(m, locales.Trf("stickers.packname_success", title))
	return err
}

func repositionSticker(m *telegram.NewMessage) error {
	posStr := strings.TrimSpace(m.Args())
	if posStr == "" || !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.reposition_usage"))
		return err
	}

	pos, err := strconv.Atoi(posStr)
	if err != nil {
		_, err = eOR(m, locales.Tr("stickers.reposition_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	sticker := reply.Sticker()
	if sticker == nil {
		_, err = eOR(m, locales.Tr("stickers.not_sticker"))
		return err
	}

	_, err = client.StickersChangeStickerPosition(docToInputDoc(sticker), int32(pos))
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.reposition_error", err.Error()))
		return err
	}
	_, err = eOR(m, locales.Trf("stickers.reposition_success", pos))
	return err
}

func removeSticker(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.rmstick_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	sticker := reply.Sticker()
	if sticker == nil {
		_, err = eOR(m, locales.Tr("stickers.not_sticker"))
		return err
	}

	_, err = client.StickersRemoveStickerFromSet(docToInputDoc(sticker))
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.rmstick_error", err.Error()))
		return err
	}
	_, err = eOR(m, locales.Tr("stickers.rmstick_success"))
	return err
}

func favSticker(m *telegram.NewMessage) error {
	if !m.IsReply() {
		_, err := eOR(m, locales.Tr("stickers.favstick_usage"))
		return err
	}

	reply, err := m.GetReplyMessage()
	if err != nil || reply == nil {
		_, err = eOR(m, locales.Tr("stickers.reply_error"))
		return err
	}

	sticker := reply.Sticker()
	if sticker == nil {
		_, err = eOR(m, locales.Tr("stickers.not_sticker"))
		return err
	}

	unfave := strings.TrimSpace(m.Args()) == "remove"
	_, err = client.MessagesFaveSticker(docToInputDoc(sticker), unfave)
	if err != nil {
		_, err = eOR(m, locales.Trf("stickers.favstick_error", err.Error()))
		return err
	}

	if unfave {
		_, err = eOR(m, locales.Tr("stickers.favstick_removed"))
	} else {
		_, err = eOR(m, locales.Tr("stickers.favstick_added"))
	}
	return err
}

func setPackCmd(m *telegram.NewMessage) error {
	name := strings.TrimSpace(m.Args())
	if name == "" {
		current := getPackShortName()
		if current == "" {
			_, err := eOR(m, locales.Tr("stickers.no_user_pack"))
			return err
		}
		_, err := eOR(m, locales.Trf("stickers.current_pack", current, current), &telegram.SendOptions{ParseMode: "HTML"})
		return err
	}
	setPackShortName(name)
	_, err := eOR(m, locales.Trf("stickers.pack_set", name, name), &telegram.SendOptions{ParseMode: "HTML"})
	return err
}

func loadStickersModule() {
	handlers := []*Handler{
		{ModuleName: "Stickers", Command: "kang", Description: "Kang a sticker to your pack", Func: kangSticker},
		{ModuleName: "Stickers", Command: "pkang", Description: "Kang an entire sticker pack", Func: kangPack},
		{ModuleName: "Stickers", Command: "setpack", Description: "Set or view current sticker pack", Func: setPackCmd},
		{ModuleName: "Stickers", Command: "listpacks", Description: "List all your sticker packs", Func: listPacksCmd},
		{ModuleName: "Stickers", Command: "setemoji", Description: "Change sticker emoji", Func: setStickerEmoji, DisAllowSudos: true},
		{ModuleName: "Stickers", Command: "packname", Description: "Rename your sticker pack", Func: renamePackCmd, DisAllowSudos: true},
		{ModuleName: "Stickers", Command: "reposition", Description: "Change sticker position", Func: repositionSticker, DisAllowSudos: true},
		{ModuleName: "Stickers", Command: "rmstick", Description: "Remove sticker from pack", Func: removeSticker, DisAllowSudos: true},
		{ModuleName: "Stickers", Command: "favstick", Description: "Add/remove sticker from favorites", Func: favSticker},
	}
	AddHandlers(handlers, client)
}
