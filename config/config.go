package config

import "time"

type Config struct {
	BaseURL             string
	APIKey              string
	PollInterval        time.Duration
	DefaultReminderHour int
}

func Default() *Config {
	return &Config{
		BaseURL:             "https://api.example.com",
		APIKey:              "YOUR_API_KEY_HERE",
		PollInterval:        5 * time.Minute,
		DefaultReminderHour: 21, // 9 PM
	}
}
