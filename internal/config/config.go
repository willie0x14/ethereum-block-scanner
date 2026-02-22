package config

import (
	"os"
	"strconv"
	"time"
)

var (
	DefaultPollInterval = 3 * time.Second
)

type Config struct {
	PollInterval time.Duration
}

func Load() *Config {

	interval := DefaultPollInterval

	// 允許用 env 覆蓋（秒）
	if v := os.Getenv("POLL_INTERVAL_SECONDS"); v != "" {
		if sec, err := strconv.Atoi(v); err == nil {
			interval = time.Duration(sec) * time.Second
		}
	}

	return &Config{
		PollInterval: interval,
	}
}
