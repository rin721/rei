package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// AsyncSubmitter 描述可异步提交任务的窄接口。
type AsyncSubmitter interface {
	SubmitDefault(context.Context, func()) error
}

// Config 描述日志器的最小配置。
type Config struct {
	Level  string
	Prefix string
	Writer io.Writer
}

// Logger 提供线程安全的最小日志封装。
type Logger struct {
	mu       sync.RWMutex
	level    int
	prefix   string
	writer   io.Writer
	fields   map[string]any
	executor AsyncSubmitter
}

const (
	levelDebug = iota
	levelInfo
	levelWarn
	levelError
	levelFatal
)

// New 创建一个新的日志器实例。
func New(cfg Config) (*Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	return &Logger{
		level:  level,
		prefix: cfg.Prefix,
		writer: writer,
		fields: make(map[string]any),
	}, nil
}

// Debug 输出调试级别日志。
func (l *Logger) Debug(message string) {
	l.log(levelDebug, "DEBUG", message)
}

// Info 输出信息级别日志。
func (l *Logger) Info(message string) {
	l.log(levelInfo, "INFO", message)
}

// Warn 输出告警级别日志。
func (l *Logger) Warn(message string) {
	l.log(levelWarn, "WARN", message)
}

// Error 输出错误级别日志。
func (l *Logger) Error(message string) {
	l.log(levelError, "ERROR", message)
}

// Fatal 输出致命级别日志。
//
// 为了保持库层可测试性，这个实现只记录日志，不直接退出进程。
func (l *Logger) Fatal(message string) {
	l.log(levelFatal, "FATAL", message)
}

// With 返回带附加字段的新日志器。
func (l *Logger) With(fields map[string]any) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	merged := make(map[string]any, len(l.fields)+len(fields))
	for key, value := range l.fields {
		merged[key] = value
	}
	for key, value := range fields {
		merged[key] = value
	}

	return &Logger{
		level:    l.level,
		prefix:   l.prefix,
		writer:   l.writer,
		fields:   merged,
		executor: l.executor,
	}
}

// Sync 在 writer 支持同步时刷新底层缓冲。
func (l *Logger) Sync() error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	type syncer interface {
		Sync() error
	}

	if writer, ok := l.writer.(syncer); ok {
		return writer.Sync()
	}

	return nil
}

// Reload 以原子方式应用新的日志配置。
func (l *Logger) Reload(cfg Config) error {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = level
	l.prefix = cfg.Prefix
	if cfg.Writer != nil {
		l.writer = cfg.Writer
	}

	return nil
}

// SetExecutor 设置可选的异步执行器。
func (l *Logger) SetExecutor(executor AsyncSubmitter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.executor = executor
}

func (l *Logger) log(requiredLevel int, levelName, message string) {
	l.mu.RLock()
	if requiredLevel < l.level {
		l.mu.RUnlock()
		return
	}

	line := formatLogLine(levelName, l.prefix, message, l.fields)
	writer := l.writer
	executor := l.executor
	l.mu.RUnlock()

	writeFn := func() {
		_, _ = io.WriteString(writer, line)
	}

	if executor != nil {
		if err := executor.SubmitDefault(context.Background(), writeFn); err == nil {
			return
		}
	}

	writeFn()
}

func formatLogLine(levelName, prefix, message string, fields map[string]any) string {
	var builder strings.Builder

	builder.WriteString(time.Now().UTC().Format(time.RFC3339))
	builder.WriteString(" level=")
	builder.WriteString(levelName)
	if prefix != "" {
		builder.WriteString(" prefix=")
		builder.WriteString(prefix)
	}
	builder.WriteString(" msg=")
	builder.WriteString(strconvQuote(message))

	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			builder.WriteString(" ")
			builder.WriteString(key)
			builder.WriteString("=")
			builder.WriteString(strconvQuote(fmt.Sprint(fields[key])))
		}
	}

	builder.WriteString("\n")
	return builder.String()
}

func parseLevel(level string) (int, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "", "info":
		return levelInfo, nil
	case "debug":
		return levelDebug, nil
	case "warn", "warning":
		return levelWarn, nil
	case "error":
		return levelError, nil
	case "fatal":
		return levelFatal, nil
	default:
		return 0, fmt.Errorf("unsupported logger level %q", level)
	}
}

func strconvQuote(value string) string {
	return fmt.Sprintf("%q", value)
}
