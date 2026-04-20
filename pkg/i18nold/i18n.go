package i18n

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Config 描述 I18n 配置。
type Config struct {
	DefaultLocale  string
	FallbackLocale string
	LocaleDir      string
}

// Manager 提供 locale 加载、选择和消息查找能力。
type Manager struct {
	mu       sync.RWMutex
	cfg      Config
	messages map[string]map[string]string
}

// New 创建一个新的 I18n 管理器。
func New(cfg Config) (*Manager, error) {
	normalized, messages, err := loadMessages(cfg)
	if err != nil {
		return nil, err
	}

	return &Manager{
		cfg:      normalized,
		messages: messages,
	}, nil
}

// Localize 返回指定 locale 的本地化消息。
func (m *Manager) Localize(locale, key string, data map[string]any) (string, error) {
	m.mu.RLock()
	cfg := m.cfg
	messages := m.messages
	m.mu.RUnlock()

	selected := pickLocale(messages, cfg, locale)
	message, ok := messages[selected][key]
	if !ok {
		if fallbackMessage, exists := messages[cfg.FallbackLocale][key]; exists {
			message = fallbackMessage
		} else {
			return "", fmt.Errorf("message %q not found for locale %q", key, selected)
		}
	}

	if !strings.Contains(message, "{{") {
		return message, nil
	}

	tmpl, err := template.New(key).Option("missingkey=zero").Parse(message)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// PickLocale 根据偏好选择最合适的 locale。
func (m *Manager) PickLocale(preferred string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return pickLocale(m.messages, m.cfg, preferred)
}

// Reload 使用新配置重新加载 locale 目录。
func (m *Manager) Reload(cfg Config) error {
	normalized, messages, err := loadMessages(cfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = normalized
	m.messages = messages
	return nil
}

func loadMessages(cfg Config) (Config, map[string]map[string]string, error) {
	if cfg.DefaultLocale == "" {
		cfg.DefaultLocale = "zh-CN"
	}
	if cfg.FallbackLocale == "" {
		cfg.FallbackLocale = "en-US"
	}
	if cfg.LocaleDir == "" {
		return Config{}, nil, errors.New("locale dir is required")
	}

	files, err := filepath.Glob(filepath.Join(cfg.LocaleDir, "*.yaml"))
	if err != nil {
		return Config{}, nil, err
	}
	if len(files) == 0 {
		return Config{}, nil, errors.New("no locale files found")
	}

	messages := make(map[string]map[string]string, len(files))
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return Config{}, nil, err
		}

		values := make(map[string]string)
		if err := yaml.Unmarshal(content, &values); err != nil {
			return Config{}, nil, err
		}

		locale := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		messages[locale] = values
	}

	return cfg, messages, nil
}

func pickLocale(messages map[string]map[string]string, cfg Config, preferred string) string {
	preferred = strings.TrimSpace(strings.Split(preferred, ",")[0])
	if _, ok := messages[preferred]; ok {
		return preferred
	}

	for locale := range messages {
		if strings.EqualFold(locale, preferred) {
			return locale
		}
	}

	if _, ok := messages[cfg.DefaultLocale]; ok {
		return cfg.DefaultLocale
	}
	if _, ok := messages[cfg.FallbackLocale]; ok {
		return cfg.FallbackLocale
	}

	for locale := range messages {
		return locale
	}

	return cfg.DefaultLocale
}
