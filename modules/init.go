package modules

import (
	"NovaUserbot/config"
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"NovaUserbot/utils"
	"fmt"
	"regexp"
	"time"

	"github.com/go-redis/redis/v8"

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
	cmdPrefix string
)

func InitTgClients() (*telegram.Client, error) {
	if err := locales.Init(Db); err != nil {
		logger.Warn("i18n init:", err)
	}

	gogramLogger := telegram.NewLogger(telegram.LogInfo, telegram.LoggerConfig{
		Output: logger.GetLogWriter(),
	})

	var err error
	client, err = telegram.NewClient(telegram.ClientConfig{
		AppID:         cfg.ApiId,
		AppHash:       cfg.ApiHash,
		LogLevel:      telegram.LogInfo,
		Logger:        gogramLogger,
		StringSession: cfg.StringSession,
		MemorySession: true,
		SessionName:   "asstub",
	})
	if err != nil {
		return nil, err
	}

	if err = client.Connect(); err != nil {
		return nil, err
	}
	if err = client.Start(); err != nil {
		return nil, err
	}

	tgbot, err = telegram.NewClient(telegram.ClientConfig{
		AppID:       cfg.ApiId,
		AppHash:     cfg.ApiHash,
		LogLevel:    telegram.LogInfo,
		Logger:      gogramLogger,
		SessionName: "asstbot",
		Session:     "asstbot.db",
	})
	if err != nil {
		return nil, err
	}

	if err = tgbot.Connect(); err != nil {
		return nil, err
	}
	if err = tgbot.LoginBot(cfg.Token); err != nil {
		return nil, err
	}

	user, err := client.GetMe()
	if err != nil || user == nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	ubId = user.ID
	logger.Infof("Logged in as %s (%d)", user.Username, ubId)

	bot, err := tgbot.GetMe()
	if err != nil || bot == nil {
		return nil, fmt.Errorf("failed to get bot info: %v", err)
	}
	tbotId = bot.ID
	logger.Infof("Bot logged in as %s (%d)", bot.Username, tbotId)

	cmdPrefix = db.Get("CMD_HANDLER")
	if cmdPrefix == "" {
		cmdPrefix = "."
	}
	logger.Infof("Command prefix: %s", cmdPrefix)

	logger.SetTelegramSender(&TgLogSender{})

	loadAllModules(client)
	return client, nil
}

func loadAllModules(c *telegram.Client) {
	defer LoadHelpHandler(c)

	LoadSystemModule(c)

	LoadAdminModule(c)
	LoadSudoModule(c)
	LoadGbanHandler(c)
	LoadBanGuardModule(c)

	LoadMyinfo(c)
	LoadProfileModule(c)
	LoadPmAssistantHandler(c)
	LoadAFKModule(c)
	LoadStoriesModule(c)

	LoadSearchModule(c)
	LoadIMDBModule(c)
	LoadUnsplashModule(c)

	LoadAudioToolsModule(c)
	LoadImageToolsModule(c)
	LoadMediaToolsModule(c)
	LoadFilesModule(c)
	LoadFileShareModule(c)
	LoadPasteModule(c)
	LoadGDriveModule(c)

	LoadDbCmds(c)
	LoadLanguageModule(c)
	LoadLoggingModule(c)
	LoadTagLogger(c)
	LoadRemindersModule(c)
	LoadSpeedTestModule(c)

	LoadChatBotHandler(c)

	LoadShellHandler(c)
	LoadUpdaterModule(c)

	logger.Info("All modules loaded")
	logger.Startup(fmt.Sprintf(locales.Tr("system.started"), time.Since(startTime).String()))
}

func AddHandlers(handlers []*Handler, c *telegram.Client) {
	for _, h := range handlers {
		AddHandler(h, c)
	}
}

func AddHandler(h *Handler, c *telegram.Client) {
	if h.Command != "" {

		prefix := cmdPrefix
		if prefix == "" {
			prefix = "."
		}
		escapedPrefix := regexp.QuoteMeta(prefix)
		c.On(fmt.Sprintf("message:%s%s( (.*)|$)", escapedPrefix, h.Command), h.Func, telegram.FilterFunc(func(m *telegram.NewMessage) bool {
			return m.Sender.ID == ubId || (utils.IsIn64Array(sudoers, m.Sender.ID) && !h.DisAllowSudos)
		}))
	}
	if h.Description != "" {
		HelpMap[h.ModuleName] = append(HelpMap[h.ModuleName], *h)
	}
}

func loadSudoers() {
	sudos, err := db.SMembers("SUDOS")
	if err != nil || len(sudos) == 0 {
		return
	}
	for _, sudo := range sudos {
		sudoId := utils.StringToInt64(sudo)
		if sudoId != 0 {
			sudoers = append(sudoers, sudoId)
		}
	}
	logger.Infof("Loaded %d sudoers", len(sudoers))
}

func InitUb() {
	if err := logger.Init(); err != nil {
		logger.Warn("Logger init warning:", err)
	}

	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		logger.Fatal("Config error:", err)
	}

	Db, err = db.InitDB(cfg.DbUrl)
	if err != nil {
		logger.Fatal("Database error:", err)
	}

	loadSudoers()

	client, err := InitTgClients()
	if err != nil {
		logger.Fatal("Telegram error:", err)
	}

	client.Idle()

	if tgbot != nil {
		tgbot.Stop()
	}
	if client != nil {
		client.Stop()
	}
	db.Close()
	logger.Shutdown("Userbot stopped")
}
