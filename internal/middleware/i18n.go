package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rin721/rei/types/constants"
)

// I18n 选择请求使用的 locale 并写入上下文。
func I18n(cfg MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		localeHeader := cfg.LocaleHeader
		if localeHeader == "" {
			localeHeader = constants.HeaderAcceptLanguage
		}

		locale := constants.DefaultLocale
		if cfg.I18n != nil {
			locale = pickLocale(cfg.I18n, c.GetHeader(localeHeader))
		}

		c.Set(constants.ContextKeyLocale, locale)
		c.Header("Content-Language", locale)
		c.Next()
	}
}

func pickLocale(translator interface {
	IsSupported(string) bool
	GetDefaultLanguage() string
}, acceptLanguage string) string {
	defaultLocale := translator.GetDefaultLanguage()
	if defaultLocale == "" {
		defaultLocale = constants.DefaultLocale
	}

	for _, part := range strings.Split(acceptLanguage, ",") {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		if idx := strings.Index(candidate, ";"); idx >= 0 {
			candidate = strings.TrimSpace(candidate[:idx])
		}
		if candidate == "" {
			continue
		}
		if translator.IsSupported(candidate) {
			return candidate
		}

		base := strings.ToLower(candidate)
		switch {
		case strings.HasPrefix(base, "zh") && translator.IsSupported(constants.DefaultLocale):
			return constants.DefaultLocale
		case strings.HasPrefix(base, "en") && translator.IsSupported(constants.FallbackLocale):
			return constants.FallbackLocale
		}
	}

	return defaultLocale
}
