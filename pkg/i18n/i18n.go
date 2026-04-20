// Package i18n 提供国际化(i18n)支持
// 基于 go-i18n/v2 库,支持多语言消息翻译
// 设计目标:
// - 统一的多语言管理
// - 灵活的语言选择
// - 简单易用的 API
// - 支持热加载翻译文件
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// I18n 定义国际化管理接口
// 提供消息翻译和语言管理功能
// 为什么使用接口:
// - 抽象实现细节
// - 便于测试和 mock
// - 支持不同的实现方式
type I18n interface {
	// T 翻译消息
	// 这是最常用的方法,根据消息 ID 和语言返回翻译后的文本
	// 参数:
	//   lang: 目标语言(如 "zh-CN", "en-US")
	//   messageID: 消息 ID,对应翻译文件中的 key
	//   templateData: 可选的模板数据,用于填充消息中的占位符
	// 返回:
	//   string: 翻译后的消息文本
	// 使用示例:
	//   msg := i18n.T("zh-CN", "welcome.message")
	//   msg := i18n.T("zh-CN", "user.greeting", map[string]interface{}{"Name": "Alice"})
	T(lang string, messageID string, templateData ...map[string]interface{}) string

	// MustT 翻译消息,如果翻译失败则 panic
	// 用于必须成功翻译的场景
	// 参数同 T
	// 返回:
	//   string: 翻译后的消息文本
	// 使用场景:
	//   - 关键错误消息
	//   - 系统级提示
	MustT(lang string, messageID string, templateData ...map[string]interface{}) string

	// IsSupported 检查语言是否被支持
	// 参数:
	//   lang: 语言代码
	// 返回:
	//   bool: true 表示支持,false 表示不支持
	IsSupported(lang string) bool

	// GetDefaultLanguage 获取默认语言
	// 返回:
	//   string: 默认语言代码
	GetDefaultLanguage() string

	// LoadMessages 从指定目录加载翻译文件
	// 支持 JSON 和 YAML 格式
	// 参数:
	//   dir: 翻译文件目录
	// 返回:
	//   error: 加载失败时的错误
	LoadMessages(dir string) error
}

// Config I18n 配置
type Config struct {
	// DefaultLanguage 默认语言
	// 当无法确定用户语言时使用
	DefaultLanguage string

	// SupportedLanguages 支持的语言列表
	// 例如: ["zh-CN", "en-US", "ja-JP"]
	SupportedLanguages []string

	// MessagesDir 翻译文件目录
	// 包含各语言的翻译文件
	// 例如: ./locales/
	//   - zh-CN.yaml
	//   - en-US.yaml
	//   - ja-JP.yaml
	MessagesDir string
}

// i18nImpl 实现 I18n 接口
// 使用 go-i18n/v2 库进行消息翻译
type i18nImpl struct {
	// bundle 消息包,管理所有语言的翻译
	bundle *i18n.Bundle

	// defaultLanguage 默认语言
	defaultLanguage string

	// supportedLanguages 支持的语言集合
	// 使用 map 提高查询效率
	supportedLanguages map[string]bool
}

// New 创建一个新的 I18n 实例
// 参数:
//
//	cfg: I18n 配置
//
// 返回:
//
//	I18n: I18n 接口实例
//	error: 创建失败时的错误
//
// 使用示例:
//
//	cfg := &i18n.Config{
//	    DefaultLanguage: "zh-CN",
//	    SupportedLanguages: []string{"zh-CN", "en-US"},
//	    MessagesDir: "./locales",
//	}
//	i18n, err := i18n.New(cfg)
func New(cfg *Config) (I18n, error) {
	// 验证配置
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = DefaultLanguage
	}

	if len(cfg.SupportedLanguages) == 0 {
		cfg.SupportedLanguages = SupportedLanguagesStringSlice
	}

	// 解析默认语言
	defaultLang, err := language.Parse(cfg.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid default language: %w", err)
	}

	// 创建 Bundle
	bundle := i18n.NewBundle(defaultLang)

	// 注册解析器
	// 支持 JSON 格式
	bundle.RegisterUnmarshalFunc(FilenameFormatJson, json.Unmarshal)
	// 支持 YAML 格式
	bundle.RegisterUnmarshalFunc(FilenameFormatYaml, yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc(FilenameFormatYml, yaml.Unmarshal)

	// 构建支持的语言集合
	supportedLangs := make(map[string]bool)
	for _, lang := range cfg.SupportedLanguages {
		supportedLangs[lang] = true
	}

	impl := &i18nImpl{
		bundle:             bundle,
		defaultLanguage:    cfg.DefaultLanguage,
		supportedLanguages: supportedLangs,
	}

	// 如果指定了消息目录,加载翻译文件
	if cfg.MessagesDir != "" {
		if err := impl.LoadMessages(cfg.MessagesDir); err != nil {
			return nil, fmt.Errorf("failed to load messages: %w", err)
		}
	}

	return impl, nil
}

// T 翻译消息
// 实现 I18n 接口
func (impl *i18nImpl) T(lang string, messageID string, templateData ...map[string]interface{}) string {
	// 如果语言不支持,使用默认语言
	if !impl.IsSupported(lang) {
		lang = impl.defaultLanguage
	}

	// 创建本地化器
	localizer := i18n.NewLocalizer(impl.bundle, lang)

	// 构建配置
	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	// 如果提供了模板数据,添加到配置中
	if len(templateData) > 0 && templateData[0] != nil {
		config.TemplateData = templateData[0]
	}

	// 翻译消息
	msg, err := localizer.Localize(config)
	if err != nil {
		// 翻译失败,返回消息 ID
		// 这样至少能让开发者知道哪个消息没有翻译
		return messageID
	}

	return msg
}

// MustT 翻译消息,失败时 panic
// 实现 I18n 接口
func (impl *i18nImpl) MustT(lang string, messageID string, templateData ...map[string]interface{}) string {
	// 如果语言不支持,使用默认语言
	if !impl.IsSupported(lang) {
		lang = impl.defaultLanguage
	}

	// 创建本地化器
	localizer := i18n.NewLocalizer(impl.bundle, lang)

	// 构建配置
	config := &i18n.LocalizeConfig{
		MessageID: messageID,
	}

	// 如果提供了模板数据,添加到配置中
	if len(templateData) > 0 && templateData[0] != nil {
		config.TemplateData = templateData[0]
	}

	// 翻译消息
	msg, err := localizer.Localize(config)
	if err != nil {
		// 翻译失败,panic
		panic(fmt.Sprintf("translation failed for message ID '%s': %v", messageID, err))
	}

	return msg
}

// IsSupported 检查语言是否被支持
// 实现 I18n 接口
func (impl *i18nImpl) IsSupported(lang string) bool {
	return impl.supportedLanguages[lang]
}

// GetDefaultLanguage 获取默认语言
// 实现 I18n 接口
func (impl *i18nImpl) GetDefaultLanguage() string {
	return impl.defaultLanguage
}

// LoadMessages 从目录加载翻译文件
// 实现 I18n 接口
// 支持的文件格式:
//   - JSON: *.json
//   - YAML: *.yaml, *.yml
//
// 文件命名规范:
//   - 语言代码.扩展名
//   - 例如: zh-CN.yaml, en-US.json
func (impl *i18nImpl) LoadMessages(dir string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("messages directory does not exist: %s", dir)
	}

	// 读取目录中的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// OnlyOnceFormatJoin 辅助函数,用于给文件格式添加点号前缀
	// 作用: 将格式名称(如 "json", "yaml")转换为文件扩展名(如 ".json", ".yaml")
	// 参数: s - 文件格式字符串(不带点号)
	// 返回: 带点号的文件扩展名字符串
	// 使用这个函数的目的:
	//   1. 避免字符串拼接的重复代码
	//   2. 统一文件扩展名的格式处理
	//   3. 与常量定义保持一致(常量不包含点号,但文件扩展名需要点号)
	OnlyOnceFormatJoin := func(s string) string {
		return fmt.Sprintf(".%s", s)
	}

	// 加载每个文件
	loaded := 0
	for _, file := range files {
		// 跳过目录
		if file.IsDir() {
			continue
		}

		// 获取文件名和扩展名
		filename := file.Name()
		ext := filepath.Ext(filename)

		// 只处理支持的格式
		if ext != OnlyOnceFormatJoin(FilenameFormatJson) && ext != OnlyOnceFormatJoin(FilenameFormatYaml) && ext != OnlyOnceFormatJoin(FilenameFormatYml) {
			continue
		}

		// 加载翻译文件
		fullPath := filepath.Join(dir, filename)
		if _, err := impl.bundle.LoadMessageFile(fullPath); err != nil {
			return fmt.Errorf("failed to load message file %s: %w", filename, err)
		}

		loaded++
	}

	// 检查是否至少加载了一个文件
	if loaded == 0 {
		return fmt.Errorf("no message files found in directory: %s", dir)
	}

	return nil
}

// Default 创建一个使用默认配置的 I18n 实例
// 默认配置:
//   - 默认语言: zh-CN
//   - 支持语言: zh-CN, en-US
//   - 不加载消息文件(需要手动调用 LoadMessages)
//
// 返回:
//
//	I18n: I18n 接口实例
//
// 使用场景:
//   - 快速开始
//   - 测试环境
func Default() I18n {
	impl, err := New(&Config{
		DefaultLanguage:    DefaultLanguage,
		SupportedLanguages: SupportedLanguagesStringSlice,
	})
	if err != nil {
		// 使用默认配置不应该失败
		panic(fmt.Sprintf("failed to create default i18n: %v", err))
	}
	return impl
}
