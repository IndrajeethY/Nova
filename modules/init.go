package modules

import (
	"NovaUserbot/config"
	"NovaUserbot/db"
	"fmt"
	"io"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func setupLogger() error {
	if _, err := os.Stat("bot_logs.json"); err == nil {
		err = os.Remove("bot_logs.json")
		if err != nil {
			return fmt.Errorf("failed to remove existing log file: %v", err)
		}
	}
	logFile := &lumberjack.Logger{
		Filename:   "bot_logs.json",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFormatter(&log.JSONFormatter{
		PrettyPrint: true,
	})
	log.SetLevel(log.InfoLevel)

	return nil
}

func InitUb() {
	var err error
	if err = setupLogger(); err != nil {
		fmt.Println("Error setting up logger:", err)
	}
	log.Println("Logger set up")
	cfg, err = config.LoadConfig()
	if err != nil {
		log.Println("Error loading config:", err)
	}
	log.Println("Config loaded")
	Db, err = db.InitDB(cfg.DbUrl)
	if err != nil {
		log.Println("Error initializing database:", err)
	}
	log.Println("Database initialized")
	if sudos, err := Db.SMembers(Db.Context(), "SUDOS").Result(); err != nil {
		for _, sudo := range sudos {
			sudoId, _ := strconv.ParseInt(sudo, 10, 64)
			sudoers = append(sudoers, sudoId)
		}
		log.Println("Error fetching sudoers:", err)
	}
	log.Println("Loaded sudoers:", len(sudoers))
	if client, err := InitTgClients(); err != nil {
		log.Println("Error initializing Telegram clients:", err)
	} else {
		log.Println("Telegram clients initialized")
		client.Idle()
		tgbot.Stop()
		client.Stop()
		log.Println("Userbot stopped")
	}
}
