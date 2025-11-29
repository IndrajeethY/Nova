package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type Reminder struct {
	ID           uint      `json:"id"`
	MessageLink  string    `json:"message_link"`
	ReminderText string    `json:"reminder_text"`
	RemindAt     time.Time `json:"remind_at"`
	CreatedAt    time.Time `json:"created_at"`
}

const maxReminders = 25

var (
	reminderMutex          sync.Mutex
	reminderCheckerRunning = false
	reminderCheckerStop    chan bool
	reminderIDCounter      uint = 0
)

func parseDurationString(input string) (time.Duration, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	re := regexp.MustCompile(`^(\d+)([dhms])`)

	var totalDuration time.Duration
	remaining := input

	for len(remaining) > 0 {
		matches := re.FindStringSubmatch(remaining)
		if len(matches) < 3 {
			return 0, fmt.Errorf("invalid duration format: %s", input)
		}

		value, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", matches[1])
		}

		unit := matches[2]
		switch unit {
		case "d":
			totalDuration += time.Duration(value) * 24 * time.Hour
		case "h":
			totalDuration += time.Duration(value) * time.Hour
		case "m":
			totalDuration += time.Duration(value) * time.Minute
		case "s":
			totalDuration += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("invalid time unit: %s", unit)
		}

		remaining = remaining[len(matches[0]):]
	}

	if totalDuration == 0 {
		return 0, fmt.Errorf("duration cannot be zero")
	}

	return totalDuration, nil
}

func formatDurationHuman(d time.Duration) string {
	if d < 0 {
		return "overdue"
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}

func getReminders() ([]Reminder, error) {
	data := db.Get("REMINDERS")
	if data == "" {
		return []Reminder{}, nil
	}

	var reminders []Reminder
	if err := json.Unmarshal([]byte(data), &reminders); err != nil {
		return nil, err
	}
	return reminders, nil
}

func saveReminders(reminders []Reminder) error {
	if len(reminders) == 0 {
		return db.Del("REMINDERS")
	}

	data, err := json.Marshal(reminders)
	if err != nil {
		return err
	}
	return db.Set("REMINDERS", string(data))
}

func createReminder(messageLink, reminderText string, remindAt time.Time) error {
	reminders, err := getReminders()
	if err != nil {
		return err
	}

	reminderMutex.Lock()
	reminderIDCounter++
	newID := reminderIDCounter
	reminderMutex.Unlock()

	reminder := Reminder{
		ID:           newID,
		MessageLink:  messageLink,
		ReminderText: reminderText,
		RemindAt:     remindAt,
		CreatedAt:    time.Now(),
	}

	reminders = append(reminders, reminder)
	return saveReminders(reminders)
}

func deleteReminderByID(reminderID uint) error {
	reminders, err := getReminders()
	if err != nil {
		return err
	}

	var newReminders []Reminder
	for _, r := range reminders {
		if r.ID != reminderID {
			newReminders = append(newReminders, r)
		}
	}

	return saveReminders(newReminders)
}

func remindCommand(m *telegram.NewMessage) error {
	args := strings.Fields(m.Args())
	if len(args) < 2 {
		_, err := eOR(m, locales.Tr("reminders.usage"))
		return err
	}

	duration, err := parseDurationString(args[0])
	if err != nil {
		_, err := eOR(m, locales.Tr("reminders.invalid_duration"))
		return err
	}

	if duration > 30*24*time.Hour {
		_, err := eOR(m, locales.Tr("reminders.too_long"))
		return err
	}

	if duration < 1*time.Minute {
		_, err := eOR(m, locales.Tr("reminders.too_short"))
		return err
	}

	reminders, err := getReminders()
	if err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	if len(reminders) >= maxReminders {
		_, err := eOR(m, locales.Tr("reminders.limit_reached"))
		return err
	}

	reminderText := strings.Join(args[1:], " ")
	if len(reminderText) > 500 {
		_, err := eOR(m, locales.Tr("reminders.text_too_long"))
		return err
	}

	remindAt := time.Now().Add(duration)
	messageLink := msgLink(m)

	if err := createReminder(messageLink, reminderText, remindAt); err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	_, err = eOR(m, fmt.Sprintf(locales.Tr("reminders.created"), formatDurationHuman(duration), reminderText))
	return err
}

func remindersListCommand(m *telegram.NewMessage) error {
	reminders, err := getReminders()
	if err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	if len(reminders) == 0 {
		_, err := eOR(m, locales.Tr("reminders.none"))
		return err
	}

	text := locales.Tr("reminders.list_header") + "\n\n"

	for i, reminder := range reminders {
		timeUntil := time.Until(reminder.RemindAt)

		var timeStr string
		if timeUntil < 0 {
			timeStr = "pending"
		} else {
			timeStr = formatDurationHuman(timeUntil)
		}

		if reminder.MessageLink != "" {
			text += fmt.Sprintf("%d. <a href='%s'>%s</a> - %s\n",
				i+1, reminder.MessageLink, reminder.ReminderText, timeStr)
		} else {
			text += fmt.Sprintf("%d. %s - %s\n",
				i+1, reminder.ReminderText, timeStr)
		}
	}

	_, err = eOR(m, text)
	return err
}

func delReminderCommand(m *telegram.NewMessage) error {
	args := strings.TrimSpace(m.Args())
	if args == "" {
		_, err := eOR(m, locales.Tr("reminders.usage_delreminder"))
		return err
	}

	index, err := strconv.Atoi(args)
	if err != nil || index < 1 {
		_, err := eOR(m, locales.Tr("reminders.invalid_index"))
		return err
	}

	reminders, err := getReminders()
	if err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	if index > len(reminders) {
		_, err := eOR(m, locales.Tr("reminders.not_found"))
		return err
	}

	reminderID := reminders[index-1].ID
	if err := deleteReminderByID(reminderID); err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	_, err = eOR(m, locales.Tr("reminders.deleted"))
	return err
}

func clearRemindersCommand(m *telegram.NewMessage) error {
	if err := db.Del("REMINDERS"); err != nil {
		_, err := eOR(m, locales.Tr("reminders.error"))
		return err
	}

	_, err := eOR(m, locales.Tr("reminders.cleared"))
	return err
}

func StartReminderChecker() {
	reminderMutex.Lock()
	if reminderCheckerRunning {
		reminderMutex.Unlock()
		return
	}
	reminderCheckerRunning = true
	reminderCheckerStop = make(chan bool)
	reminderMutex.Unlock()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				checkAndTriggerReminders()
			case <-reminderCheckerStop:
				logger.Info("Reminder checker stopped")
				return
			}
		}
	}()
}

func StopReminderChecker() {
	reminderMutex.Lock()
	defer reminderMutex.Unlock()

	if !reminderCheckerRunning {
		return
	}

	reminderCheckerRunning = false
	close(reminderCheckerStop)
}

func checkAndTriggerReminders() {
	if tgbot == nil {
		return
	}

	reminders, err := getReminders()
	if err != nil || len(reminders) == 0 {
		return
	}

	now := time.Now()
	var updatedReminders []Reminder

	for _, reminder := range reminders {
		if now.After(reminder.RemindAt) {
			sendReminderNotification(reminder)
		} else {
			updatedReminders = append(updatedReminders, reminder)
		}
	}

	saveReminders(updatedReminders)
}

func sendReminderNotification(reminder Reminder) {
	if tgbot == nil {
		return
	}

	ownerMention := fmt.Sprintf("<a href='tg://user?id=%d'>Owner</a>", ubId)

	var text string
	if reminder.MessageLink != "" {
		text = fmt.Sprintf(locales.Tr("reminders.notification_with_link"), reminder.ReminderText, reminder.MessageLink)
	} else {
		text = fmt.Sprintf(locales.Tr("reminders.notification"), reminder.ReminderText)
	}

	text = ownerMention + "\n" + text

	logChatStr := db.Get("LOG_CHAT")
	logChat := utils.StringToInt64(logChatStr)
	if logChat == 0 {
		logChat = ubId
	}

	peer, err := tgbot.GetSendablePeer(logChat)
	if err != nil {
		logger.Errorf("Failed to get peer for reminder %d: %v", reminder.ID, err)
		return
	}

	_, err = tgbot.SendMessage(peer, text, &telegram.SendOptions{
		ParseMode: "HTML",
	})

	if err != nil {
		logger.Errorf("Failed to send reminder %d: %v", reminder.ID, err)
	}
}

func LoadRemindersModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "remind", Func: remindCommand, Description: "Set a reminder (e.g., .remind 1h30m Buy groceries)", ModuleName: "Reminders", DisAllowSudos: true},
		{Command: "reminders", Func: remindersListCommand, Description: "List your reminders", ModuleName: "Reminders", DisAllowSudos: true},
		{Command: "delreminder", Func: delReminderCommand, Description: "Delete a reminder by index", ModuleName: "Reminders", DisAllowSudos: true},
		{Command: "clearreminders", Func: clearRemindersCommand, Description: "Clear all your reminders", ModuleName: "Reminders", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)

	StartReminderChecker()
}
