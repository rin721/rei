package app

import "os"

func resolveConfigPath() string {
	if value, ok := os.LookupEnv("APP_CONFIG_PATH"); ok && value != "" {
		return value
	}
	return defaultConfigPath
}
