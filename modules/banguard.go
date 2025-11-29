package modules

import (
	"NovaUserbot/db"
	"NovaUserbot/locales"
	"NovaUserbot/logger"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

type BanGuardConfig struct {
	Limit    int   `json:"limit"`
	Duration int64 `json:"duration"`
	Enabled  bool  `json:"enabled"`
}

type BanCache struct {
	sync.Mutex
	cache     map[int64]map[int64]int
	timestamps map[int64]map[int64]time.Time
}

var banCache = BanCache{
	cache:     make(map[int64]map[int64]int),
	timestamps: make(map[int64]map[int64]time.Time),
}

func addUserToCache(chatID int64, adminID int64, banLimit int, duration int64) bool {
	banCache.Lock()
	defer banCache.Unlock()

	now := time.Now()

	if banCache.cache[chatID] == nil {
		banCache.cache[chatID] = make(map[int64]int)
	}
	if banCache.timestamps[chatID] == nil {
		banCache.timestamps[chatID] = make(map[int64]time.Time)
	}

	if lastTime, exists := banCache.timestamps[chatID][adminID]; exists {
		if now.Sub(lastTime) > time.Duration(duration)*time.Second {
			banCache.cache[chatID][adminID] = 0
		}
	}

	banCache.cache[chatID][adminID]++
	banCache.timestamps[chatID][adminID] = now
	return banCache.cache[chatID][adminID] >= banLimit
}

func resetUserBanCount(chatID int64, adminID int64) {
	banCache.Lock()
	defer banCache.Unlock()

	if banCache.cache[chatID] != nil {
		banCache.cache[chatID][adminID] = 0
	}
}

func getBanGuardConfig(chatID int64) *BanGuardConfig {
	data := db.Get(fmt.Sprintf("BANGUARD:%d", chatID))
	if data == "" {
		return nil
	}

	var config BanGuardConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil
	}
	return &config
}

func setBanGuardConfig(chatID int64, config *BanGuardConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return db.Set(fmt.Sprintf("BANGUARD:%d", chatID), string(data))
}

func deleteBanGuardConfig(chatID int64) error {
	return db.Del(fmt.Sprintf("BANGUARD:%d", chatID))
}

func setBanGuardLimit(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("banguard.groups_only"))
		return err
	}

	args := strings.Fields(m.Args())
	if len(args) < 2 {
		_, err := eOR(m, locales.Tr("banguard.usage_gconfig"))
		return err
	}

	duration, err := time.ParseDuration(args[0])
	if err != nil {
		_, err := eOR(m, locales.Tr("banguard.invalid_duration"))
		return err
	}

	userLimit, err := strconv.Atoi(args[1])
	if err != nil || userLimit <= 0 {
		_, err := eOR(m, locales.Tr("banguard.invalid_limit"))
		return err
	}

	config := &BanGuardConfig{
		Limit:    userLimit,
		Duration: int64(duration.Seconds()),
		Enabled:  true,
	}

	if err := setBanGuardConfig(m.ChatID(), config); err != nil {
		_, err := eOR(m, locales.Tr("banguard.config_error"))
		return err
	}

	_, err = eOR(m, fmt.Sprintf(locales.Tr("banguard.limits_set"), duration, userLimit))
	return err
}

func toggleBanGuard(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("banguard.groups_only"))
		return err
	}

	args := strings.ToLower(strings.TrimSpace(m.Args()))
	if args != "on" && args != "off" {
		_, err := eOR(m, locales.Tr("banguard.usage_gtoggle"))
		return err
	}

	config := getBanGuardConfig(m.ChatID())

	if args == "on" {
		if config == nil {
			config = &BanGuardConfig{
				Limit:    5,
				Duration: 10,
				Enabled:  true,
			}
		} else {
			config.Enabled = true
		}

		if err := setBanGuardConfig(m.ChatID(), config); err != nil {
			_, err := eOR(m, locales.Tr("banguard.config_error"))
			return err
		}

		if config.Limit == 5 && config.Duration == 10 {
			_, err := eOR(m, locales.Tr("banguard.enabled_default"))
			return err
		}
		_, err := eOR(m, locales.Tr("banguard.enabled"))
		return err
	}

	if config == nil {
		_, err := eOR(m, locales.Tr("banguard.not_configured"))
		return err
	}

	if err := deleteBanGuardConfig(m.ChatID()); err != nil {
		_, err := eOR(m, locales.Tr("banguard.config_error"))
		return err
	}

	_, err := eOR(m, locales.Tr("banguard.disabled"))
	return err
}

func banGuardStatus(m *telegram.NewMessage) error {
	if !m.IsGroup() {
		_, err := eOR(m, locales.Tr("banguard.groups_only"))
		return err
	}

	config := getBanGuardConfig(m.ChatID())
	if config == nil || !config.Enabled {
		_, err := eOR(m, locales.Tr("banguard.status_disabled"))
		return err
	}

	duration := time.Duration(config.Duration) * time.Second
	_, err := eOR(m, fmt.Sprintf(locales.Tr("banguard.status_enabled"), config.Limit, duration))
	return err
}

func UserJoinHandle(p *telegram.ParticipantUpdate) error {
	if !p.IsBanned() && !p.IsKicked() {
		return nil
	}

	chatID := p.ChatID()
	actorID := p.ActorID()

	if actorID == ubId || actorID == 0 {
		return nil
	}

	config := getBanGuardConfig(chatID)
	if config == nil || !config.Enabled {
		return nil
	}

	exceeded := addUserToCache(chatID, actorID, config.Limit, config.Duration)
	if !exceeded {
		return nil
	}

	_, err := p.Demote()
	if err != nil {
		logger.Errorf("BanGuard: Failed to demote admin %d in chat %d: %v", actorID, chatID, err)
		return nil
	}

	banCount := config.Limit
	resetUserBanCount(chatID, actorID)

	actorName := "Unknown"
	if p.Actor != nil {
		actorName = p.Actor.FirstName
		if p.Actor.LastName != "" {
			actorName += " " + p.Actor.LastName
		}
	}

	chatTitle := "Unknown"
	if p.Channel != nil {
		chatTitle = p.Channel.Title
	}

	peer, err := client.GetSendablePeer(chatID)
	if err == nil {
		client.SendMessage(peer, fmt.Sprintf(locales.Tr("banguard.user_demoted"), actorID, actorName), &telegram.SendOptions{ParseMode: "HTML"})
	}

	logMessage(fmt.Sprintf(locales.Tr("banguard.log_demoted"), chatTitle, chatID, actorID, actorName, actorID, banCount))

	logger.Infof("BanGuard: Demoted admin %d (%s) in chat %d for exceeding ban limit", actorID, actorName, chatID)
	return nil
}

func LoadBanGuardModule(c *telegram.Client) {
	handlers := []*Handler{
		{Command: "gconfig", Func: setBanGuardLimit, Description: "Set BanGuard limits (duration limit)", ModuleName: "BanGuard", DisAllowSudos: true},
		{Command: "gtoggle", Func: toggleBanGuard, Description: "Toggle BanGuard on/off", ModuleName: "BanGuard", DisAllowSudos: true},
		{Command: "gstatus", Func: banGuardStatus, Description: "Check BanGuard status", ModuleName: "BanGuard", DisAllowSudos: true},
	}
	AddHandlers(handlers, c)

	c.On(telegram.OnParticipant, UserJoinHandle)
}
