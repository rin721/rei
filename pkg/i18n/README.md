# I18n 国际化工具包

提供简单易用的多语言翻译功能,基于 [go-i18n/v2](https://github.com/nicksnyder/go-i18n) 库。

## ✨ 功能特性

- 🌍 **多语言支持** - 支持任意数量的语言
- 📝 **灵活的文件格式** - 支持 JSON 和 YAML 翻译文件
- 🔧 **消息模板** - 支持占位符和变量替换
- 🔄 **自动回退** - 翻译不存在时自动使用默认语言
- ⚡ **高性能** - 翻译查询使用内存 map,极速响应
- 🔒 **线程安全** - 可在多个 goroutine 中并发使用

## 📦 安装

```bash
go get github.com/nicksnyder/go-i18n/v2
go get golang.org/x/text
```

## 🚀 快速开始

### 1. 创建翻译文件

创建 `locales` 目录,添加语言文件:

**locales/zh-CN.yaml:**

```yaml
welcome.message: 欢迎使用我们的应用
user.greeting: 你好, {{.Name}}!
error.user_not_found: 用户不存在
error.invalid_params: 参数错误
success.user_created: 用户创建成功
```

**locales/en-US.yaml:**

```yaml
welcome.message: Welcome to our application
user.greeting: Hello, {{.Name}}!
error.user_not_found: User not found
error.invalid_params: Invalid parameters
success.user_created: User created successfully
```

### 2. 初始化 I18n

```go
package main

import (
    "log"
    "github.com/rei0721/go-scaffold/pkg/i18n"
)

func main() {
    // 创建配置
    cfg := &i18n.Config{
        DefaultLanguage:    "zh-CN",
        SupportedLanguages: []string{"zh-CN", "en-US"},
        MessagesDir:        "./locales",
    }

    // 创建 I18n 实例
    i18n, err := i18n.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // 使用翻译
    msg := i18n.T("zh-CN", "welcome.message")
    fmt.Println(msg) // 输出: 欢迎使用我们的应用

    msg = i18n.T("en-US", "welcome.message")
    fmt.Println(msg) // 输出: Welcome to our application
}
```

### 3. 使用模板变量

```go
// 带变量的消息翻译
msg := i18n.T("zh-CN", "user.greeting", map[string]interface{}{
    "Name": "张三",
})
fmt.Println(msg) // 输出: 你好, 张三!

msg = i18n.T("en-US", "user.greeting", map[string]interface{}{
    "Name": "Alice",
})
fmt.Println(msg) // 输出: Hello, Alice!
```

## 🔧 在 Gin 框架中使用

### 创建中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/rei0721/go-scaffold/pkg/i18n"
)

// I18n 中间件提取并存储用户的语言偏好
func I18n(i18n i18n.I18n) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Accept-Language 头部获取语言
        lang := c.GetHeader("Accept-Language")

        // 如果语言不支持,使用默认语言
        if lang == "" || !i18n.IsSupported(lang) {
            lang = i18n.GetDefaultLanguage()
        }

        // 存储到上下文
        c.Set("lang", lang)
        c.Set("i18n", i18n)

        c.Next()
    }
}
```

### 在处理器中使用

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/rei0721/go-scaffold/pkg/i18n"
)

type UserHandler struct {
    i18n i18n.I18n
}

func NewUserHandler(i18n i18n.I18n) *UserHandler {
    return &UserHandler{i18n: i18n}
}

func (h *UserHandler) GetUser(c *gin.Context) {
    // 获取语言
    lang, _ := c.Get("lang")
    langStr := lang.(string)

    // 查询用户...
    user, err := h.service.GetUser(id)
    if err != nil {
        // 使用翻译的错误消息
        msg := h.i18n.T(langStr, "error.user_not_found")
        c.JSON(http.StatusNotFound, gin.H{"error": msg})
        return
    }

    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    lang, _ := c.Get("lang")
    langStr := lang.(string)

    // 创建用户...
    if err := h.service.CreateUser(req); err != nil {
        msg := h.i18n.T(langStr, "error.invalid_params")
        c.JSON(http.StatusBadRequest, gin.H{"error": msg})
        return
    }

    // 返回成功消息
    msg := h.i18n.T(langStr, "success.user_created")
    c.JSON(http.StatusOK, gin.H{"message": msg})
}
```

## 📚 API 文档

### Config

配置结构体:

```go
type Config struct {
    DefaultLanguage    string   // 默认语言
    SupportedLanguages []string // 支持的语言列表
    MessagesDir        string   // 翻译文件目录
}
```

### I18n 接口

```go
type I18n interface {
    // T 翻译消息
    T(lang string, messageID string, templateData ...map[string]interface{}) string

    // MustT 翻译消息,失败时 panic
    MustT(lang string, messageID string, templateData ...map[string]interface{}) string

    // IsSupported 检查语言是否被支持
    IsSupported(lang string) bool

    // GetDefaultLanguage 获取默认语言
    GetDefaultLanguage() string

    // LoadMessages 从目录加载翻译文件
    LoadMessages(dir string) error
}
```

### 创建实例

```go
// New 创建 I18n 实例
func New(cfg *Config) (I18n, error)

// Default 使用默认配置创建实例
func Default() I18n
```

## 📋 翻译文件格式

### JSON 格式

```json
{
  "welcome.message": "欢迎使用我们的应用",
  "user.greeting": "你好, {{.Name}}!",
  "user.info": "用户 {{.Name}} 已经注册了 {{.Days}} 天"
}
```

### YAML 格式

```yaml
welcome.message: 欢迎使用我们的应用
user.greeting: 你好, {{.Name}}!
user.info: 用户 {{.Name}} 已经注册了 {{.Days}} 天
```

## 🎯 最佳实践

### 1. 使用有意义的消息 ID

✅ **推荐:**

```go
i18n.T(lang, "error.user_not_found")
i18n.T(lang, "success.user_created")
i18n.T(lang, "validation.email_invalid")
```

❌ **避免:**

```go
i18n.T(lang, "err1")
i18n.T(lang, "msg_001")
i18n.T(lang, "text123")
```

### 2. 分组管理消息 ID

```yaml
# 错误消息
error.user_not_found: 用户不存在
error.invalid_params: 参数错误
error.unauthorized: 未授权

# 成功消息
success.user_created: 用户创建成功
success.profile_updated: 资料更新成功

# 验证消息
validation.email_invalid: 邮箱格式不正确
validation.password_too_short: 密码太短
```

### 3. 始终提供默认语言翻译

确保所有消息在默认语言中都有定义,其他语言缺失时会自动回退。

### 4. 使用模板而不是字符串拼接

✅ **推荐:**

```go
i18n.T(lang, "user.greeting", map[string]interface{}{
    "Name": userName,
})
```

❌ **避免:**

```go
msg := "Hello, " + userName + "!"
```

### 5. 在配置文件中管理路径

```yaml
# config.yaml
i18n:
  default: zh-CN
  supported:
    - zh-CN
    - en-US
    - ja-JP
  messages_dir: ./locales
```

## 🌍 支持的语言代码

常用语言代码:

| 语言       | 代码  | 说明                     |
| ---------- | ----- | ------------------------ |
| 简体中文   | zh-CN | Chinese (Simplified)     |
| 英语(美国) | en-US | English (United States)  |
| 英语(英国) | en-GB | English (United Kingdom) |
| 日语       | ja-JP | Japanese                 |
| 韩语       | ko-KR | Korean                   |
| 法语       | fr-FR | French                   |
| 德语       | de-DE | German                   |
| 西班牙语   | es-ES | Spanish                  |

更多语言代码请参考: [IETF Language Tag](https://en.wikipedia.org/wiki/IETF_language_tag)

## ⚡ 性能说明

- **Bundle 创建**: 应用启动时创建一次 (毫秒级)
- **消息加载**: 启动时从文件加载 (毫秒级)
- **翻译查询**: 内存 map 查询 (纳秒级)
- **Localizer**: 每次创建开销很小 (微秒级)

总体来说,I18n 对运行时性能的影响可以忽略不计。

## 🔒 线程安全

`i18n.Bundle` 是线程安全的,可以在多个 goroutine 中并发调用 `T()` 方法,不需要额外的同步措施。

## ❗ 错误处理

### T() 方法

翻译失败时返回消息 ID:

```go
msg := i18n.T("zh-CN", "non.existent.key")
// msg = "non.existent.key" (返回 ID 本身)
```

### MustT() 方法

翻译失败时会 panic,适合关键消息:

```go
msg := i18n.MustT("zh-CN", "critical.error")
// 如果翻译失败,程序会 panic
```

**建议:** 一般情况使用 `T()`,关键系统消息使用 `MustT()`。

## 📁 项目结构示例

```
project/
├── locales/              # 翻译文件目录
│   ├── zh-CN.yaml       # 简体中文
│   ├── en-US.yaml       # 英语
│   └── ja-JP.yaml       # 日语
├── internal/
│   ├── middleware/
│   │   └── i18n.go      # I18n 中间件
│   └── handler/
│       └── user.go      # 使用 I18n 的处理器
├── pkg/
│   └── i18n/            # I18n 包
└── main.go
```

## 🔗 参考资料

- [go-i18n 官方文档](https://github.com/nicksnyder/go-i18n)
- [IETF 语言标签标准](https://en.wikipedia.org/wiki/IETF_language_tag)
- [Go text/template 语法](https://pkg.go.dev/text/template)

## 📄 许可证

本项目采用与主项目相同的许可证。
