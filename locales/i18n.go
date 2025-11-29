package locales

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yml
var localeFiles embed.FS

type Translations struct {
	mu          sync.RWMutex
	languages   map[string]map[string]interface{}
	defaultLang string
	db          *redis.Client
}

var (
	instance *Translations
	once     sync.Once
)

func GetInstance() *Translations {
	once.Do(func() {
		instance = &Translations{
			languages:   make(map[string]map[string]interface{}),
			defaultLang: "en",
		}
	})
	return instance
}

func Init(db *redis.Client) error {
	t := GetInstance()
	t.db = db

	entries, err := localeFiles.ReadDir("locales")
	if err != nil {
		log.Printf("Warning: Could not read locales directory: %v", err)
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			if err := t.loadLanguageFile(entry.Name()); err != nil {
				log.Printf("Warning: Could not load language file %s: %v", entry.Name(), err)
			}
		}
	}

	log.Printf("Loaded %d languages", len(t.languages))
	return nil
}

func (t *Translations) loadLanguageFile(filename string) error {
	data, err := localeFiles.ReadFile("locales/" + filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var langData map[string]interface{}
	if err := yaml.Unmarshal(data, &langData); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	langCode := strings.TrimSuffix(filename, ".yml")

	t.mu.Lock()
	t.languages[langCode] = langData
	t.mu.Unlock()

	return nil
}

func (t *Translations) GetUserLanguage(userID int64) string {
	if t.db == nil {
		return t.defaultLang
	}

	lang, err := t.db.Get(context.Background(), fmt.Sprintf("USER_LANG_%d", userID)).Result()
	if err != nil || lang == "" {
		lang, err = t.db.Get(context.Background(), "BOT_LANGUAGE").Result()
		if err != nil || lang == "" {
			return t.defaultLang
		}
	}
	return lang
}

func (t *Translations) SetUserLanguage(userID int64, lang string) error {
	if t.db == nil {
		return nil
	}
	return t.db.Set(context.Background(), fmt.Sprintf("USER_LANG_%d", userID), lang, 0).Err()
}

func (t *Translations) SetGlobalLanguage(lang string) error {
	if t.db == nil {
		return nil
	}
	return t.db.Set(context.Background(), "BOT_LANGUAGE", lang, 0).Err()
}

func (t *Translations) Get(lang, key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langData, exists := t.languages[lang]
	if !exists {
		langData = t.languages[t.defaultLang]
		if langData == nil {
			return key
		}
	}

	parts := strings.Split(key, ".")
	var current interface{} = langData

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			if lang != t.defaultLang {
				return t.Get(t.defaultLang, key)
			}
			return key
		}
	}

	if str, ok := current.(string); ok {
		return str
	}

	if lang != t.defaultLang {
		return t.Get(t.defaultLang, key)
	}
	return key
}

func (t *Translations) GetForUser(userID int64, key string) string {
	lang := t.GetUserLanguage(userID)
	return t.Get(lang, key)
}

func Tr(key string) string {
	return GetInstance().Get(GetInstance().defaultLang, key)
}

func TrUser(userID int64, key string) string {
	return GetInstance().GetForUser(userID, key)
}

func TrLang(lang, key string) string {
	return GetInstance().Get(lang, key)
}

func GetAvailableLanguages() []string {
	t := GetInstance()
	t.mu.RLock()
	defer t.mu.RUnlock()

	langs := make([]string, 0, len(t.languages))
	for lang := range t.languages {
		langs = append(langs, lang)
	}
	return langs
}

func GetLanguageName(code string) string {
	return GetInstance().Get(code, "language.name")
}
