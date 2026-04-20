package i18n

// DefaultLanguage 默认语言
// 当无法从请求中获取语言时使用
const DefaultLanguage = "zh-CN"

const (
	// LanguageHeader HTTP 头部中语言字段的名称
	// 客户端应该在请求头中设置此字段来指定期望的语言
	// 例如: Accept-Language: zh-CN
	LanguageHeader = "Accept-Language"

	// LanguageEnglish 英语(美国)语言代码
	// 遵循 BCP 47 标准,用于标识美式英语
	LanguageEnglish = "en-US"

	// LanguageChinese 简体中文语言代码
	// 遵循 BCP 47 标准,用于标识中国大陆使用的简体中文
	LanguageChinese = "zh-CN"

	// LanguageJapanese 日语语言代码
	// 遵循 BCP 47 标准,用于标识日本使用的日语
	LanguageJapanese = "ja-JP"

	// FilenameFormatJson JSON 文件格式标识
	// 用于 go-i18n 注册 JSON 格式的消息文件解析器
	FilenameFormatJson = "json"

	// FilenameFormatYaml YAML 文件格式标识
	// 用于 go-i18n 注册 YAML 格式的消息文件解析器
	FilenameFormatYaml = "yaml"

	// FilenameFormatYml YML 文件格式标识
	// YAML 的另一种常见扩展名,功能与 yaml 相同
	FilenameFormatYml = "yml"
)

// SupportedLanguages 支持的语言列表
// 用于快速验证语言是否被支持
var SupportedLanguages = map[string]bool{
	LanguageChinese:  true, // 简体中文
	LanguageEnglish:  true, // 英语(美国)
	LanguageJapanese: true, // 日语
}

// SupportedLanguagesStringSlice 支持的语言列表(字符串切片形式)
// 与 SupportedLanguages map 提供相同的语言支持,但以切片形式存储
// 使用场景:
//   - 作为 Config.SupportedLanguages 的默认值
//   - 需要遍历所有支持的语言时
//   - 需要保持语言顺序时
//
// 注意: 与 SupportedLanguages map 相比,切片不适合频繁的查找操作
var SupportedLanguagesStringSlice = []string{LanguageChinese, LanguageEnglish}
