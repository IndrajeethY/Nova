package modules

import (
	"NovaUserbot/utils"
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

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
	if Db.SIsMember(context.Background(), "APPROVED_USERS", m.Sender.ID).Val() {
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
		_, _ = m.Reply(fmt.Sprintf("You've reached your message limit. You will be temporarily blocked from messaging %s.", client.Me().FirstName))
		peer, _ := m.Client.GetSendablePeer(userID)
		_, err := m.Client.ContactsBlock(false, peer)
		if err != nil {
			return logMessage(fmt.Sprintf("Error blocking user %d: %s", userID, err))
		}
		return nil
	}
	messageCounts[userID] = count + 1
	countMutex.Unlock()

	var combinedPrompt string
	var prompt string
	responseMutex.Lock()
	lastResponse := lastResponses[userID]
	senderInfo := fmt.Sprintf("Sender: %s (@%s)", m.Sender.FirstName, m.Sender.Username)
	promptVal := Db.Get(context.Background(), "PM_AI_PROMT").Val()
	if promptVal == "" {
		prompt = fmt.Sprintf("Act as a personal messaging assistant for %s and he/she is your owner and your purpose is to serve him/her, responding on his behalf with professionalism and wisdom. Encourage users to keep conversations short and detailed, limiting them to three messages starting from their initial message. After three messages from a user, inform them that they've reached their message limit and will be temporarily blocked from messaging.", client.Me().FirstName)
	} else {
		prompt = promptVal
	}
	if count == 0 {
		combinedPrompt = fmt.Sprintf("%s\n%s\nUser's message: %s", prompt, senderInfo, m.Text())
	} else {
		combinedPrompt = fmt.Sprintf("%s\n%s\nPrevious AI response: %s\nUser's message: %s", prompt, senderInfo, lastResponse, m.Text())
	}
	responseMutex.Unlock()

	result, err := utils.ProcessGemini("", combinedPrompt)
	if err != nil {
		log.Error("Error processing message:", err)
		return err
	}

	_, err = m.Reply(result)
	if err != nil {
		log.Error("Error sending reply:", err)
		return err
	}
	responseMutex.Lock()
	lastResponses[userID] = result
	responseMutex.Unlock()

	return nil
}

func ApproveUser(m *telegram.NewMessage) error {
	userID, name := ExtractUser(m)
	if userID == 0 {
		_, err := eOR(m, "Invalid user ID.")
		return err
	}
	if Db.SIsMember(context.Background(), "APPROVED_USERS", userID).Val() {
		_, err := eOR(m, fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> is already approved to pm.", userID, name))
		return err
	} else {
		err := Db.SAdd(context.Background(), "APPROVED_USERS", userID).Err()
		if err != nil {
			_, err = eOR(m, "Error approving user.")
			return err
		}
		_, err = eOR(m, fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> approved to pm.", userID, name))
		return err
	}
}

func DisapproveUser(m *telegram.NewMessage) error {
	userId, name := ExtractUser(m)
	if userId == 0 {
		_, err := eOR(m, "Invalid user ID.")
		return err
	}
	if !Db.SIsMember(context.Background(), "APPROVED_USERS", userId).Val() {
		_, err := eOR(m, fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> is not approved.", userId, name))
		return err
	}
	err := Db.SRem(context.Background(), "APPROVED_USERS", userId).Err()
	if err != nil {
		_, err = eOR(m, "Error disapproving user.")
		return err
	}
	_, err = eOR(m, fmt.Sprintf("User <a href='tg://user?id=%d'>%s</a> disapproved.", userId, name))
	return err
}

func ApprovedUsers(m *telegram.NewMessage) error {
	config, err := Db.SMembers(context.Background(), "APPROVED_USERS").Result()
	if err != nil {
		_, err = eOR(m, "Error fetching approved users.")
		return err
	}
	msg, _ := eOR(m, "<code>Fetching approved users...</code>")
	output := "<b>Approved users:</b>\n"
	for _, id := range config {
		user, err := m.Client.GetUser(utils.StringToInt64(id))
		if err != nil {
			log.Error("Error getting user info:", err)
			continue
		}
		output += fmt.Sprintf("<a href='tg://user?id=%d'>%s</a>\n", utils.StringToInt64(id), user.FirstName+" "+user.LastName)
	}
	_, err = msg.Edit(output)
	return err
}

func SetPromt(m *telegram.NewMessage) error {
	prompt := m.Args()
	if prompt == "" {
		_, err := eOR(m, "Usage: .setprompt <prompt>")
		return err
	}
	err := Db.Set(context.Background(), "PM_AI_PROMT", prompt, 0).Err()
	if err != nil {
		_, err = eOR(m, "Error setting prompt.")
		return err
	}
	_, err = eOR(m, "Prompt set successfully.")
	return err
}

func LoadPmAssistantHandler(c *telegram.Client) {
	handlers := []*Handler{
		{
			ModuleName:  "Pm Permit",
			Command:     "ap",
			Description: "Approve a user to bypass the assistant prompt",
			Func:        ApproveUser,
		},
		{
			ModuleName:  "Pm Permit",
			Command:     "dap",
			Description: "Disapprove a user to receive the assistant prompt",
			Func:        DisapproveUser,
		},
		{
			ModuleName:  "Pm Permit",
			Command:     "approved",
			Description: "List approved users",
			Func:        ApprovedUsers,
		},
		{
			ModuleName:  "Pm Permit",
			Command:     "setprompt",
			Description: "Set the prompt for the pm assistant",
			Func:        SetPromt,
		},
	}
	AddHandlers(handlers, c)
	c.On("message", OnPrivateMessage)
}
