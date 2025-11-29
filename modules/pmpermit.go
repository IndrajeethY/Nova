package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"fmt"
	"sync"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	messageCounts = make(map[int64]int)
	lastResponses = make(map[int64]string)
	countMutex    sync.Mutex
	responseMutex sync.Mutex
)

func OnPrivateMessage(m *telegram.NewMessage) error {
	if !m.IsPrivate() || m.Sender.Bot || m.Sender.Contact {
		return nil
	}

	if db.SIsMember("APPROVED_USERS", m.Sender.ID) {
		return nil
	}

	peerinfo, _ := m.Client.MessagesGetPeerSettings(&telegram.InputPeerUser{
		UserID:     m.Sender.ID,
		AccessHash: m.Sender.AccessHash,
	})
	if peerinfo != nil && peerinfo.Settings != nil && !peerinfo.Settings.BlockContact {
		return nil
	}

	userID := m.Sender.ID
	countMutex.Lock()
	count := messageCounts[userID]
	if count >= 3 {
		countMutex.Unlock()
		m.Reply(fmt.Sprintf(locales.Tr("pm_permit.message_limit"), client.Me().FirstName))
		peer, _ := m.Client.GetSendablePeer(userID)
		m.Client.ContactsBlock(false, peer)
		return nil
	}
	messageCounts[userID] = count + 1
	countMutex.Unlock()

	prompt := db.Get("PM_AI_PROMT")
	if prompt == "" {
		prompt = fmt.Sprintf("Act as a personal messaging assistant for %s. Encourage users to keep conversations short. After three messages, inform them they've reached their limit.", client.Me().FirstName)
	}

	responseMutex.Lock()
	lastResponse := lastResponses[userID]
	senderInfo := fmt.Sprintf("Sender: %s (@%s)", m.Sender.FirstName, m.Sender.Username)

	var combinedPrompt string
	if count == 0 {
		combinedPrompt = fmt.Sprintf("%s\n%s\nUser's message: %s", prompt, senderInfo, m.Text())
	} else {
		combinedPrompt = fmt.Sprintf("%s\n%s\nPrevious AI response: %s\nUser's message: %s", prompt, senderInfo, lastResponse, m.Text())
	}
	responseMutex.Unlock()

	result, err := utils.ProcessGemini("", combinedPrompt)
	if err != nil {
		logger.Error("PM AI error:", err)
		return err
	}

	m.Reply(result)
	responseMutex.Lock()
	lastResponses[userID] = result
	responseMutex.Unlock()

	return nil
}

func ApproveUser(m *telegram.NewMessage) error {
	userID, name := ExtractUser(m)
	if userID == 0 {
		_, err := eOR(m, locales.Tr("pm_permit.invalid_user"))
		return err
	}

	if db.SIsMember("APPROVED_USERS", userID) {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("pm_permit.already_approved"), userID, name))
		return err
	}

	if err := db.SAdd("APPROVED_USERS", userID); err != nil {
		_, err = eOR(m, locales.Tr("pm_permit.approve_error"))
		return err
	}

	_, err := eOR(m, fmt.Sprintf(locales.Tr("pm_permit.approved"), userID, name))
	return err
}

func DisapproveUser(m *telegram.NewMessage) error {
	userId, name := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, locales.Tr("pm_permit.invalid_user"))
		return err
	}

	if !db.SIsMember("APPROVED_USERS", userId) {
		_, err := eOR(m, fmt.Sprintf(locales.Tr("pm_permit.not_approved"), userId, name))
		return err
	}

	if err := db.SRem("APPROVED_USERS", userId); err != nil {
		_, err = eOR(m, locales.Tr("pm_permit.disapprove_error"))
		return err
	}

	_, err := eOR(m, fmt.Sprintf(locales.Tr("pm_permit.disapproved"), userId, name))
	return err
}

func ApprovedUsers(m *telegram.NewMessage) error {
	users, err := db.SMembers("APPROVED_USERS")
	if err != nil {
		_, err = eOR(m, locales.Tr("pm_permit.fetch_error"))
		return err
	}

	msg, _ := eOR(m, locales.Tr("pm_permit.fetching"))
	output := locales.Tr("pm_permit.list_header") + "\n"

	for _, id := range users {
		user, err := m.Client.GetUser(utils.StringToInt64(id))
		if err != nil {
			continue
		}
		output += fmt.Sprintf("<a href='tg://user?id=%s'>%s</a>\n", id, user.FirstName+" "+user.LastName)
	}

	_, err = msg.Edit(output)
	return err
}

func SetPromt(m *telegram.NewMessage) error {
	prompt := m.Args()
	if prompt == "" {
		_, err := eOR(m, locales.Tr("pm_permit.usage_prompt"))
		return err
	}

	if err := db.Set("PM_AI_PROMT", prompt); err != nil {
		_, err = eOR(m, locales.Tr("pm_permit.prompt_error"))
		return err
	}

	_, err := eOR(m, locales.Tr("pm_permit.prompt_set"))
	return err
}

func LoadPmAssistantHandler(c *telegram.Client) {
	handlers := []*Handler{
		{ModuleName: "Pm Permit", Command: "ap", Description: "Approve a user", Func: ApproveUser},
		{ModuleName: "Pm Permit", Command: "dap", Description: "Disapprove a user", Func: DisapproveUser},
		{ModuleName: "Pm Permit", Command: "approved", Description: "List approved users", Func: ApprovedUsers},
		{ModuleName: "Pm Permit", Command: "setprompt", Description: "Set PM assistant prompt", Func: SetPromt},
	}
	AddHandlers(handlers, c)
	c.On("message", OnPrivateMessage)
}
