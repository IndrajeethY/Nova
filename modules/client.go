package modules

import (
	"NovaUserbot/config"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"

	"NovaUserbot/utils"

	"github.com/amarnathcjd/gogram/telegram"
)

var (
	client    *telegram.Client
	tgbot     *telegram.Client
	ubId      int64
	tbotId    int64
	sudoers   []int64
	startTime = time.Now()
	cfg       *config.Config
	Db        *redis.Client
)

const NovaVersion = "1.1.0"

func InitTgClients() (*telegram.Client, error) {
	client, _ = telegram.NewClient(telegram.ClientConfig{
		AppID:       cfg.ApiID,
		AppHash:     cfg.ApiHash,
		LogLevel:    telegram.LogInfo,
		SessionName: "asstub",
	})
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start userbot: %w", err)
	}

	tgbot, _ = telegram.NewClient(telegram.ClientConfig{
		AppID:       cfg.ApiID,
		AppHash:     cfg.ApiHash,
		LogLevel:    telegram.LogInfo,
		SessionName: "asstbot",
		Session:     "asstbot.db",
	})
	if err := tgbot.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect bot: %w", err)
	}
	if err := tgbot.LoginBot(cfg.BotToken); err != nil {
		return nil, fmt.Errorf("failed to login bot: %w", err)
	}

	user, _ := client.GetMe()
	log.Printf("Userbot logged in as @%s (%d)", user.Username, user.ID)
	ubId = user.ID

	bot, _ := tgbot.GetMe()
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
		pattern := fmt.Sprintf("message:%s%s( (.*)|$)", cmdPrefix, h.Command)
		c.On(pattern, h.Func, telegram.FilterFunc(func(m *telegram.NewMessage) bool {
			return m.Sender.ID == ubId || (utils.IsIn64Array(sudoers, m.Sender.ID) && !h.DisAllowSudos)
		}))
	}
	if h.Description != "" {
		HelpMap[h.ModuleName] = append(HelpMap[h.ModuleName], *h)
	}
}
