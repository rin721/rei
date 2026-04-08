package config

import (
	"os"
	"regexp"
)

var envPattern = regexp.MustCompile(`\$\{([A-Z0-9_]+)(?::([^}]*))?\}`)

// ExpandEnvPlaceholders 展开 `${VAR:default}` 形式的环境变量占位。
func ExpandEnvPlaceholders(content []byte) []byte {
	return envPattern.ReplaceAllFunc(content, func(match []byte) []byte {
		parts := envPattern.FindSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		key := string(parts[1])
		if value, ok := os.LookupEnv(key); ok {
			return []byte(value)
		}
		if len(parts) >= 3 {
			return parts[2]
		}
		return []byte{}
	})
}
