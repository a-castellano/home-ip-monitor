package config

import (
	"cmp"
	"errors"
	"os"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	redisconfig "github.com/a-castellano/go-types/redis"
)

// Config struct contians required config variables
type Config struct {
	ISPName        string // home-ip-monitor will send new IP values to be updated if associated ISP is the same than this value
	UpdateQueue    string // This will be the queue used to send IP changes
	NotifyQueue    string // This will be the queue used to notify IP or ISP changes
	RedisConfig    *redisconfig.Config
	RabbitmqConfig *rabbitmqconfig.Config
}

// NewConfig checks if required env variables are present, returns config instance
func NewConfig() (*Config, error) {
	config := Config{}
	var redisConfigErr, rabbitmqConfigErr error

	// Retrieve ISPName
	config.ISPName = cmp.Or(os.Getenv("ISP_NAME"), "no_set")

	if config.ISPName == "no_set" {
		return nil, errors.New("env variable ISP_NAME must be set")
	}

	// Retrieve UpdateQueue name, default is home-ip-monitor-updates
	config.UpdateQueue = cmp.Or(os.Getenv("UPDATE_QUEUE_NAME"), "home-ip-monitor-updates")

	// Retrieve NotifyQueue name, default is home-ip-monitor-updates
	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")

	// Set RedisConfig and RabbitmqConfig
	config.RedisConfig, redisConfigErr = redisconfig.NewConfig()
	if redisConfigErr != nil {
		return nil, redisConfigErr
	}

	config.RabbitmqConfig, rabbitmqConfigErr = rabbitmqconfig.NewConfig()
	if rabbitmqConfigErr != nil {
		return nil, rabbitmqConfigErr
	}

	return &config, nil
}
