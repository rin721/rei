// Package i18n 提供国际化(internationalization)支持
//
// # 概述
//
// i18n 包封装了 go-i18n/v2 库,提供简单易用的多语言翻译功能。
// 支持从 JSON 和 YAML 文件加载翻译,支持消息模板和占位符替换。
//
// # 功能特性
//
//   - 多语言支持:支持任意数量的语言
//   - 灵活的翻译文件格式:支持 JSON 和 YAML
//   - 消息模板:支持占位符替换
//   - 默认语言回退:当翻译不存在时自动使用默认语言
//   - 简单的 API:只需几行代码即可集成
//
// # 快速开始
//
// 创建 I18n 实例:
//
//	cfg := &i18n.Config{
//	    DefaultLanguage: "zh-CN",
//	    SupportedLanguages: []string{"zh-CN", "en-US", "ja-JP"},
//	    MessagesDir: "./locales",
//	}
//	i18n, err := i18n.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 翻译消息:
//
//	// 简单翻译
//	msg := i18n.T("zh-CN", "welcome.message")
//
//	// 带模板数据的翻译
//	msg := i18n.T("zh-CN", "user.greeting", map[string]interface{}{
//	    "Name": "Alice",
//	})
//
// # 翻译文件格式
//
// 支持 JSON 格式 (zh-CN.json):
//
//	{
//	  "welcome.message": "欢迎使用我们的应用",
//	  "user.greeting": "你好, {{.Name}}!"
//	}
//
// 支持 YAML 格式 (zh-CN.yaml):
//
//	welcome.message: 欢迎使用我们的应用
//	user.greeting: 你好, {{.Name}}!
//
// # 文件命名规范
//
// 翻译文件应该以语言代码命名:
//   - zh-CN.json 或 zh-CN.yaml - 简体中文
//   - en-US.json 或 en-US.yaml - 英语(美国)
//   - ja-JP.json 或 ja-JP.yaml - 日语
//
// # 在 Gin 中使用
//
// 创建中间件提取语言:
//
//	func I18nMiddleware(i18n i18n.I18n) gin.HandlerFunc {
//	    return func(c *gin.Context) {
//	        // 从 Accept-Language 头部获取语言
//	        lang := c.GetHeader("Accept-Language")
//	        if lang == "" || !i18n.IsSupported(lang) {
//	            lang = i18n.GetDefaultLanguage()
//	        }
//	        // 存储到上下文
//	        c.Set("lang", lang)
//	        c.Next()
//	    }
//	}
//
// 在处理器中使用:
//
//	func (h *Handler) GetUser(c *gin.Context) {
//	    lang, _ := c.Get("lang")
//	    langStr := lang.(string)
//
//	    // 获取用户...
//	    if user == nil {
//	        msg := h.i18n.T(langStr, "error.user_not_found")
//	        c.JSON(404, gin.H{"error": msg})
//	        return
//	    }
//
//	    c.JSON(200, user)
//	}
//
// # 目录结构示例
//
//	project/
//	├── locales/
//	│   ├── zh-CN.yaml
//	│   ├── en-US.yaml
//	│   └── ja-JP.yaml
//	└── main.go
//
// # 最佳实践
//
//  1. 使用有意义的消息 ID
//     - 推荐: "error.user_not_found"
//     - 避免: "err1", "msg_001"
//
//  2. 分组管理消息 ID
//     - error.xxx - 错误消息
//     - success.xxx - 成功消息
//     - validation.xxx - 验证消息
//
//  3. 始终提供默认语言的翻译
//     - 确保所有消息在默认语言中都有定义
//     - 其他语言缺失时会回退到默认语言
//
//  4. 使用模板而不是字符串拼接
//     - 推荐: T("user.greeting", map[string]interface{}{"Name": name})
//     - 避免: "Hello, " + name + "!"
//
//  5. 在配置中管理翻译文件路径
//     - 不要硬编码文件路径
//     - 使用配置文件指定 MessagesDir
//
// # 性能考虑
//
//   - Bundle 创建: 应用启动时创建一次
//   - 消息加载: 启动时加载,不会影响运行时性能
//   - 翻译查询: 使用内存中的 map,非常快速
//   - Localizer: 每次翻译创建,开销很小
//
// # 线程安全
//
// i18n.Bundle 是线程安全的,可以在多个 goroutine 中并发使用。
// 不需要额外的同步措施。
//
// # 错误处理
//
//   - T() 方法: 翻译失败时返回消息 ID
//   - MustT() 方法: 翻译失败时 panic
//   - 建议: 一般使用 T(),关键消息使用 MustT()
//
// # 参考资料
//
//   - go-i18n 文档: https://github.com/nicksnyder/go-i18n
//   - 语言代码标准: https://en.wikipedia.org/wiki/IETF_language_tag
package i18n
