package constants

// 应用级共享常量，供配置、请求链路和响应规范统一复用。
const (
	// ApplicationName 是项目的默认应用名。
	ApplicationName = "rei"
	// DefaultLocale 是默认语言。
	DefaultLocale = "zh-CN"
	// FallbackLocale 是回退语言。
	FallbackLocale = "en-US"
	// HeaderTraceID 是链路追踪请求头。
	HeaderTraceID = "X-Trace-ID"
	// HeaderAcceptLanguage 是语言协商请求头。
	HeaderAcceptLanguage = "Accept-Language"
	// ContextKeyTraceID 是上下文中的 TraceID 键。
	ContextKeyTraceID = "trace_id"
	// ContextKeyLocale 是上下文中的语言键。
	ContextKeyLocale = "locale"
	// ContextKeyUserID 是上下文中的用户标识键。
	ContextKeyUserID = "user_id"
	// ContextKeyJWTClaims 是上下文中的 JWT Claims 键。
	ContextKeyJWTClaims = "jwt_claims"
	// PhaseTag 是当前脚手架实现阶段标记。
	PhaseTag = "phase_0_7"
)
