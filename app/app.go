package app

import (
	"context"
	"log"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	"github.com/a-castellano/home-ip-monitor/ipinfo"
	"github.com/a-castellano/home-ip-monitor/monitor"
	"github.com/a-castellano/home-ip-monitor/nslookup"
)

// Monitor orchestrates the IP monitoring process by:
// 1. Initializing Redis client for data storage
// 2. Initializing RabbitMQ client for message queuing
// 3. Starting the core monitoring logic
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - requester: Interface for fetching IP information
//   - nsLookup: Interface for DNS resolution
//   - appConfig: Application configuration
//
// Returns error if any step fails
func Monitor(ctx context.Context, requester ipinfo.Requester, nsLookup nslookup.NSLookup, appConfig *config.Config) error {

	log.Print("Creating Redis client")
	// Initialize Redis client for persistent storage
	redisClient := memorydatabase.NewRedisClient(appConfig.RedisConfig)
	log.Print("Initiating Redis client")
	redisClientError := redisClient.Initiate(ctx)
	if redisClientError != nil {
		log.Print(redisClientError.Error())
		return redisClientError
	}
	// Create memory database interface for storage operations
	memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

	log.Print("Creating RabbitMQ client")
	// Initialize RabbitMQ client for message queuing
	rabbitmqClient := messagebroker.NewRabbitmqClient(appConfig.RabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	// Start the core monitoring process with all initialized components
	monitorError := monitor.Monitor(ctx, requester, nsLookup, memoryDatabase, messageBroker, appConfig)
	if monitorError != nil {
		return monitorError
	}
	log.Print("Execution finished")

	return nil
}
