package config

import (
	"cmp"
	"errors"
	"log"
	"os"

	rabbitmqconfig "github.com/a-castellano/go-types/rabbitmq"
	redisconfig "github.com/a-castellano/go-types/redis"
)

// Config struct contains required config variables
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
func NewConfig() (*Config, error) {
	config := Config{}
	var redisConfigErr, rabbitmqConfigErr error

	// Retrieve DomainName
	config.DomainName = cmp.Or(os.Getenv("DOMAIN_NAME"), "no_set")

	if config.DomainName == "no_set" {
		return nil, errors.New("env variable DOMAIN_NAME must be set")
	}
	log.Printf("Domain name has been set to \"%s\"", config.DomainName)

	// Retrieve ISPName
	config.ISPName = cmp.Or(os.Getenv("ISP_NAME"), "no_set")

	if config.ISPName == "no_set" {
		return nil, errors.New("env variable ISP_NAME must be set")
	}
	log.Printf("ISP name has been set to \"%s\"", config.ISPName)

	// Retrieve DNSServer
	config.DNSServer = cmp.Or(os.Getenv("DNS_SERVER"), "no_set")

	if config.DNSServer == "no_set" {
		return nil, errors.New("env variable DNS_SERVER must be set")
	}
	log.Printf("DNS Server has been set to \"%s\"", config.DNSServer)

	// Retrieve UpdateQueue name, default is home-ip-monitor-updates
	config.UpdateQueue = cmp.Or(os.Getenv("UPDATE_QUEUE_NAME"), "home-ip-monitor-updates")
	log.Printf("Update queue name has been set to \"%s\"", config.UpdateQueue)

	// Retrieve NotifyQueue name, default is home-ip-monitor-updates
	config.NotifyQueue = cmp.Or(os.Getenv("NOTIFY_QUEUE_NAME"), "home-ip-monitor-notifications")
	log.Printf("Notify queue name has been set to \"%s\"", config.NotifyQueue)

	// Set RedisConfig and RabbitmqConfig
	log.Print("Setting Redis Config")
	config.RedisConfig, redisConfigErr = redisconfig.NewConfig()
	if redisConfigErr != nil {
		return nil, redisConfigErr
	}

	log.Print("Setting RabbitMQ Config")
	config.RabbitmqConfig, rabbitmqConfigErr = rabbitmqconfig.NewConfig()
	if rabbitmqConfigErr != nil {
		return nil, rabbitmqConfigErr
	}

	return &config, nil
}
