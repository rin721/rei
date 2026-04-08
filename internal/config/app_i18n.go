package config

import "fmt"

// I18nConfig 描述国际化配置。
type I18nConfig struct {
	DefaultLocale  string `yaml:"default_locale" env:"I18N_DEFAULT_LOCALE"`
	FallbackLocale string `yaml:"fallback_locale" env:"I18N_FALLBACK_LOCALE"`
	LocaleDir      string `yaml:"locale_dir" env:"I18N_LOCALE_DIR"`
}

// ValidateName 返回配置域名。
func (c I18nConfig) ValidateName() string {
	return "i18n"
}

// ValidateRequired 返回该配置域是否必需。
func (c I18nConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 I18nConfig。
func (c I18nConfig) Validate() error {
	if c.DefaultLocale == "" {
		return fmt.Errorf("default_locale is required")
	}
	if c.FallbackLocale == "" {
		return fmt.Errorf("fallback_locale is required")
	}
	if c.LocaleDir == "" {
		return fmt.Errorf("locale_dir is required")
	}
	return nil
}
