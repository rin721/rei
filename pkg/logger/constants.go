package logger

type Level int8

// Level 定义日志级别
const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var LevelNames = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
	"fatal": LevelFatal,
}

func parseLevel(l string) Level {
	for k, v := range LevelNames {
		if k == l {
			return v
		}
	}
	return LevelInfo
}

const (
	DefaultLevel  = "debug"
	DefaultFormat = "console"
	DefaultOutput = "stdout"

	// Output modes 日志输出模式
	OutputStdout = "stdout" // 仅输出到标准输出（控制台）
	OutputFile   = "file"   // 仅输出到文件
	OutputBoth   = "both"   // 同时输出到文件和控制台

	// MsgLoggerReloading 日志重载中消息
	MsgLoggerReloading = "reloading logger configuration"

	// MsgLoggerReloaded 日志重载成功消息
	MsgLoggerReloaded = "logger configuration reloaded successfully"

	// ErrMsgReloadFailed 重载失败的错误消息
	ErrMsgReloadFailed = "failed to reload logger: %w"
)
