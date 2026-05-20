package modules

import (
	"NovaUserbot/config"
	"context"
	"fmt"
	"reflect"
	"runtime"
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
	cfg       *config.ConfigType
	Db        *redis.Client
)

const (
	// Version of the bot
	NovaVersion = "1.0.0"
)

func InitTgClients() (*telegram.Client, error) {
	client, _ = telegram.NewClient(telegram.ClientConfig{
		AppID:         cfg.ApiId,
		AppHash:       cfg.ApiHash,
		LogLevel:      telegram.LogInfo,
		StringSession: cfg.StringSession,
		MemorySession: true,
		SessionName:   "asstub",
	})
	if err := client.Connect(); err != nil {
		log.Println("Error connecting userbot to Telegram:", err)
	}
	if err := client.Start(); err != nil {
		log.Println("Error starting userbot:", err)
	}
	tgbot, _ = telegram.NewClient(telegram.ClientConfig{
		AppID:       cfg.ApiId,
		AppHash:     cfg.ApiHash,
		LogLevel:    telegram.LogInfo,
		SessionName: "asstbot",
		Session:     "asstbot.db",
	})
	if err := tgbot.Connect(); err != nil {
		log.Println("Error connecting bot to Telegram:", err)
		return nil, err

	}
	if err := tgbot.LoginBot(cfg.Token); err != nil {
		log.Println("Error logging in bot:", err)
		return nil, err
	}

	user, _ := client.GetMe()
	log.Println("Logged in as", user.Username)
	ubId = user.ID
	bot, _ := tgbot.GetMe()
	log.Println("Logged in as", bot.Username)
	tbotId = bot.ID

	loadAllModules(client)
	return client, nil
}

// Need a better way to load modules

func loadAllModules(client *telegram.Client) {
	modules := []func(*telegram.Client){
		LoadAdminModule,
		LoadAliveCmd,
		LoadChatBotHandler,
		LoadDbCmds,
		LoadGbanHandler,
		LoadMisc,
		LoadMyinfo,
		LoadPingHandler,
		LoadPmAssistantHandler,
		LoadShellHandler,
		LoadSudoModule,
		LoadTagLogger,
	}

	for _, load := range modules {
		name := runtime.FuncForPC(reflect.ValueOf(load).Pointer()).Name()
		log.Println("Loading module", name)
		load(client)
	}
	LoadHelpHandler(client)
	logMessage("NovaUserbot started in " + time.Since(startTime).String())
}

func AddHandlers(handlers []*Handler, client *telegram.Client) {
	for _, h := range handlers {
		AddHandler(h, client)
	}
}

func AddHandler(h *Handler, client *telegram.Client) {
	if h.Command != "" {
		cmD := Db.Get(context.Background(), "CMD_HANDLER").Val()
		if cmD == "" {
			cmD = "."
		}
		client.On(fmt.Sprintf("message:%s%s( (.*)|$)", cmD, h.Command), h.Func, telegram.FilterFunc(func(m *telegram.NewMessage) bool {
			return m.Sender.ID == ubId || (utils.IsIn64Array(sudoers, m.Sender.ID) && !h.DisAllowSudos)
		}))
	}
	if h.Description != "" {
		HelpMap[h.ModuleName] = append(HelpMap[h.ModuleName], *h)
	}
}
