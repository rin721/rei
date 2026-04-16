package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rin721/go-scaffold2/types/constants"
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
			locale = cfg.I18n.PickLocale(c.GetHeader(localeHeader))
		}

		c.Set(constants.ContextKeyLocale, locale)
		c.Header("Content-Language", locale)
		c.Next()
	}
}
