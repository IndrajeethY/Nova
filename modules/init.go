package modules

import (
	"NovaUserbot/config"
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"fmt"
	"io"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupLogger() error {
	os.Remove("bot_logs.json")
	logFile := &lumberjack.Logger{
		Filename:   "bot_logs.json",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	log.SetLevel(log.InfoLevel)
	return nil
}

func InitUb() {
	if err := setupLogger(); err != nil {
		fmt.Println("Error setting up logger:", err)
	}

	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Config loaded")

	Db, err = db.InitDB(cfg.DbURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Database connected")

	if err := locales.Init(Db); err != nil {
		log.Warnf("Failed to load locales: %v", err)
	}

	loadSudoers()

	c, err := InitTgClients()
	if err != nil {
		log.Fatalf("Failed to initialize Telegram clients: %v", err)
	}

	log.Println("NovaUserbot is running")
	c.Idle()
	tgbot.Stop()
	c.Stop()
	log.Println("NovaUserbot stopped")
}

func loadSudoers() {
	sudos, err := Db.SMembers(Db.Context(), "SUDOS").Result()
	if err != nil {
		log.Warnf("Could not load sudoers: %v", err)
		return
	}
	for _, s := range sudos {
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		sudoers = append(sudoers, id)
	}
	log.Printf("Loaded %d sudoers", len(sudoers))
}
