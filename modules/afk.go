package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type AFKData struct {
	IsAFK       bool                `json:"is_afk"`
	Reason      string              `json:"reason"`
	StartTime   time.Time           `json:"start_time"`
	MsgCount    int                 `json:"msg_count"`
	LastNotify  map[int64]time.Time `json:"last_notify"`
	MediaFileID string              `json:"media_file_id,omitempty"`
}

const afkSpamBlock = 2 * time.Minute

var (
	afkMutex sync.RWMutex
)

func getAFK() AFKData {
	raw := db.Get("AFK_DATA")
	if raw == "" {
		return AFKData{LastNotify: make(map[int64]time.Time)}
	}
	var d AFKData
	_ = json.Unmarshal([]byte(raw), &d)
	if d.LastNotify == nil {
		d.LastNotify = make(map[int64]time.Time)
	}
	return d
}

func setAFK(d AFKData) {
	j, _ := json.Marshal(d)
	_ = db.Set("AFK_DATA", string(j))
}

func fmtDur(d time.Duration) string {
	if d < 0 {
		return "just now"
	}
	if d.Hours() >= 24 {
		return fmt.Sprintf("%dd %dh", int(d.Hours()/24), int(d.Hours())%24)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	if d.Minutes() >= 1 {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func afkCommand(m *telegram.NewMessage) error {
	reason := strings.TrimSpace(m.Args())
	if reason == "" {
		reason = "No reason specified"
	}

	var fileID string
	if m.IsReply() {
		if r, err := m.GetReplyMessage(); err == nil && r.File != nil {
			fileID = r.File.FileID
		}
	}

	afkMutex.Lock()
	setAFK(AFKData{
		IsAFK:       true,
		Reason:      reason,
		StartTime:   time.Now(),
		MsgCount:    0,
		LastNotify:  map[int64]time.Time{},
		MediaFileID: fileID,
	})
	afkMutex.Unlock()

	_, err := eOR(m, fmt.Sprintf(locales.Tr("afk.now_afk"), reason))
	return err
}

func afkHandler(m *telegram.NewMessage) error {
	
	if m == nil {
		return nil
	}

	afkMutex.RLock()
	d := getAFK()
	afkMutex.RUnlock()

	if !d.IsAFK {
		return nil
	}

	
	if m.Message.Out {
		afkMutex.Lock()
		old := d
		setAFK(AFKData{IsAFK: false, LastNotify: map[int64]time.Time{}})
		afkMutex.Unlock()

		msg := fmt.Sprintf(locales.Tr("afk.auto_back"), fmtDur(time.Since(old.StartTime)), old.MsgCount)
		m.Reply(msg, &telegram.SendOptions{ParseMode: "HTML"})
		return nil
	}

	
	if !m.Message.Mentioned || m.Sender.Bot {
		return nil
	}
	afkMutex.Lock()
	d = getAFK()
	d.MsgCount++

	last, ok := d.LastNotify[m.Sender.ID]
	send := !ok || time.Since(last) > afkSpamBlock
	if send {
		d.LastNotify[m.Sender.ID] = time.Now()
	}
	setAFK(d)
	afkMutex.Unlock()

	if !send {
		return nil
	}

	msg := fmt.Sprintf(locales.Tr("afk.reply"), fmtDur(time.Since(d.StartTime)), d.Reason)

	opts := &telegram.SendOptions{ParseMode: "HTML"}
	if d.MediaFileID != "" {
		if f, err := telegram.ResolveBotFileID(d.MediaFileID); err == nil {
			opts.Media = f
		}
	}
	m.Reply(msg, opts)
	return nil
}

func LoadAFKModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "afk", Func: afkCommand, Description: "Set AFK", ModuleName: "AFK", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)

	c.On("message", afkHandler)
}
