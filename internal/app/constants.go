package app

import "time"

const (
	defaultConfigPath = "configs/config.example.yaml"
)

const (
	defaultConfigWatchInterval = 200 * time.Millisecond
	defaultShutdownTimeout     = 5 * time.Second
)
