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
		return fmt.Errorf("could not read locales directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yml") {
			if err := t.loadLanguageFile(entry.Name()); err != nil {
				log.Warnf("Could not load language file %s: %v", entry.Name(), err)
			}
		}
	}

	if t.db != nil {
		lang, err := t.db.Get(context.Background(), "BOT_LANGUAGE").Result()
		if err == nil && lang != "" {
			t.defaultLang = lang
		}
	}

	log.Printf("Loaded %d language(s), default: %s", len(t.languages), t.defaultLang)
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
		return strings.TrimRight(str, "\n")
	}

	if lang != t.defaultLang {
		return t.Get(t.defaultLang, key)
	}
	return key
}

func (t *Translations) SetGlobalLanguage(lang string) error {
	t.mu.Lock()
	t.defaultLang = lang
	t.mu.Unlock()
	if t.db != nil {
		return t.db.Set(context.Background(), "BOT_LANGUAGE", lang, 0).Err()
	}
	return nil
}

func Tr(key string) string {
	return GetInstance().Get(GetInstance().defaultLang, key)
}

func Trf(key string, args ...any) string {
	template := Tr(key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
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
