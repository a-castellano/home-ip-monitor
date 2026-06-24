package config

import (
	"cmp"
	"context"
	"errors"
	"os"

	logger "github.com/a-castellano/go-services/infra/logger"
	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	redisconfig "github.com/a-castellano/go-types/redis"
)

// Config struct contains required config variables for the home IP monitor service
type Config struct {
	DomainName     string // The domain that should be used to check if home IP values mismatch
	ISPName        string // home-ip-monitor will send new IP values to be updated if associated ISP is the same than this value
	UpdateQueue    string // This will be the queue used to send IP changes
	NotifyQueue    string // This will be the queue used to notify IP or ISP changes
	DNSServer      string // This will be the external DNS Server used to notify for checking if home IP values mismatch
	RedisConfig    *redisconfig.Config
	RabbitmqConfig *rabbitmqconfig.Config
}

// NewConfig checks if required env variables are present, returns config instance
// It validates all required environment variables and initializes Redis and RabbitMQ configurations
//
// Required environment variables:
//   - DOMAIN_NAME: Domain to verify IP against
//   - ISP_NAME: Expected ISP provider name
//   - DNS_SERVER: External DNS server for lookups
//
// Optional environment variables (with defaults):
//   - UPDATE_QUEUE_NAME: Queue for IP updates (default: "home-ip-monitor-updates")
//   - NOTIFY_QUEUE_NAME: Queue for notifications (default: "home-ip-monitor-notifications")
//
// Returns:
//   - *Config: Initialized configuration struct
//   - error: Configuration error if any required variable is missing
func NewConfig(ctx context.Context) (*Config, error) {

	log := logger.FromContext(ctx).With("operation", "NewConfig")

	config := Config{}
	var redisConfigErr, rabbitmqConfigErr error

	// Retrieve DomainName from environment
	config.DomainName = cmp.Or(os.Getenv("DOMAIN_NAME"), "no_set")

	if config.DomainName == "no_set" {
		return nil, errors.New("env variable DOMAIN_NAME must be set")
	}
	log.DebugContext(ctx, "Domain name has been set", "domain", config.DomainName)

	// Retrieve ISPName from environment
	config.ISPName = cmp.Or(os.Getenv("ISP_NAME"), "no_set")

	if config.ISPName == "no_set" {
		return nil, errors.New("env variable ISP_NAME must be set")
	}
	log.DebugContext(ctx, "ISP name has been set", "isp", config.ISPName)

	// Retrieve DNSServer from environment
	config.DNSServer = cmp.Or(os.Getenv("DNS_SERVER"), "no_set")

	if config.DNSServer == "no_set" {
		dnsError := errors.New("env variable DNS_SERVER must be set")
		log.ErrorContext(ctx, "Error configuring dns server", "error", dnsError)
		return nil, dnsError
	}
	log.DebugContext(ctx, "DNS Server has been set", "dns", config.DNSServer)

	// Retrieve UpdateQueue name, default is home-ip-monitor-updates
	config.UpdateQueue = cmp.Or(os.Getenv("UPDATE_QUEUE_NAME"), "home-ip-monitor-updates")
	log.DebugContext(ctx, "Update queue name has been set", "updatequeue", config.UpdateQueue)

	// Retrieve NotifyQueue name, default is home-ip-monitor-notifications
	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")
	log.DebugContext(ctx, "Notify queue name has been set", "notifyqueue", config.NotifyQueue)

	// Set RedisConfig and RabbitmqConfig
	log.DebugContext(ctx, "Setting Redis config")
	config.RedisConfig, redisConfigErr = redisconfig.NewConfig()
	if redisConfigErr != nil {
		log.ErrorContext(ctx, "Error setting redis config", "error", redisConfigErr)
		return nil, redisConfigErr
	}
	log.DebugContext(ctx, "Redis config has been set", "config", config.RedisConfig)

	log.DebugContext(ctx, "Setting RabbitMQ Config")
	config.RabbitmqConfig, rabbitmqConfigErr = rabbitmqconfig.NewConfig()
	if rabbitmqConfigErr != nil {
		log.ErrorContext(ctx, "Error setting RabbitMQ config", "error", rabbitmqConfigErr)
		return nil, rabbitmqConfigErr
	}
	log.DebugContext(ctx, "RabbitMQ config has been set", "config", config.RabbitmqConfig)

	return &config, nil
}
