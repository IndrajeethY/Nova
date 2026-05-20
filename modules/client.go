package modules

import (
	"NovaUserbot/config"
	"context"
	"fmt"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	client    *telegram.Client
	tgbot     *telegram.Client
	ubId      int64
	tbotId    int64
	sudoers   []int64
	sudoMu    sync.RWMutex
	helpMu    sync.RWMutex
	startTime = time.Now()
	cfg       *config.Config
	Db        *redis.Client
)

const NovaVersion = "1.1.0"

func IsSudoer(id int64) bool {
	sudoMu.RLock()
	defer sudoMu.RUnlock()
	return slices.Contains(sudoers, id)
}

func AddSudoer(id int64) {
	sudoMu.Lock()
	defer sudoMu.Unlock()
	sudoers = append(sudoers, id)
}

func RemoveSudoer(id int64) {
	sudoMu.Lock()
	defer sudoMu.Unlock()
	for i, s := range sudoers {
		if s == id {
			sudoers = append(sudoers[:i], sudoers[i+1:]...)
			return
		}
	}
}

func GetSudoersCount() int {
	sudoMu.RLock()
	defer sudoMu.RUnlock()
	return len(sudoers)
}

func SetSudoers(ids []int64) {
	sudoMu.Lock()
	defer sudoMu.Unlock()
	sudoers = ids
}

func AddHelpEntry(module string, h Handler) {
	helpMu.Lock()
	defer helpMu.Unlock()
	HelpMap[module] = append(HelpMap[module], h)
}

func GetHelpModule(module string) ([]Handler, bool) {
	helpMu.RLock()
	defer helpMu.RUnlock()
	h, ok := HelpMap[module]
	return h, ok
}

func InitTgClients() (*telegram.Client, error) {
	var err error
	client, err = telegram.NewClient(telegram.ClientConfig{
		AppID:         cfg.ApiID,
		AppHash:       cfg.ApiHash,
		LogLevel:      telegram.LogInfo,
		StringSession: cfg.StringSession,
		SessionName:   "asstub",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create userbot client: %w", err)
	}
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start userbot: %w", err)
	}

	tgbot, err = telegram.NewClient(telegram.ClientConfig{
		AppID:       cfg.ApiID,
		AppHash:     cfg.ApiHash,
		LogLevel:    telegram.LogInfo,
		SessionName: "asstbot",
		Session:     "asstbot.db",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot client: %w", err)
	}
	if err := tgbot.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect bot: %w", err)
	}
	if err := tgbot.LoginBot(cfg.BotToken); err != nil {
		return nil, fmt.Errorf("failed to login bot: %w", err)
	}

	user, err := client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get userbot info: %w", err)
	}
	log.Printf("Userbot logged in as @%s (%d)", user.Username, user.ID)
	ubId = user.ID

	bot, err := tgbot.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}
	log.Printf("Bot logged in as @%s (%d)", bot.Username, bot.ID)
	tbotId = bot.ID

	loadAllModules()
	logMessage("NovaUserbot started in " + time.Since(startTime).String())
	return client, nil
}

func AddHandlers(handlers []*Handler, c *telegram.Client) {
	for _, h := range handlers {
		AddHandler(h, c)
	}
}

func AddHandler(h *Handler, c *telegram.Client) {
	if h.Command != "" {
		cmdPrefix := Db.Get(context.Background(), "CMD_HANDLER").Val()
		if cmdPrefix == "" {
			cmdPrefix = "."
		}
		pattern := fmt.Sprintf("message:%s%s( (.*)|$)", regexp.QuoteMeta(cmdPrefix), h.Command)
		disallow := h.DisAllowSudos
		c.On(pattern, h.Func, telegram.Any(
			telegram.IsOutgoing,
			telegram.CustomFilter(func(m *telegram.NewMessage) bool {
				return IsSudoer(m.Sender.ID) && !disallow
			}),
		))
	}
	if h.Description != "" {
		AddHelpEntry(h.ModuleName, *h)
	}
}
