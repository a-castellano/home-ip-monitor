package app

import (
	"context"
	"log"
	"net/http"
	"os"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	"github.com/a-castellano/home-ip-monitor/ipinfo"
	"github.com/a-castellano/home-ip-monitor/monitor"
)

func Monitor(ctx context.Context, httpClient http.Client, appConfig *config.Config) error {

	requester := ipinfo.Realrequester{Client: httpClient}
	log.Print("Creating Redis client")
	redisClient := memorydatabase.NewRedisClient(appConfig.RedisConfig)
	log.Print("Initiating Redis client")
	redisClientError := redisClient.Initiate(ctx)
	if redisClientError != nil {
		log.Print(redisClientError.Error())
		os.Exit(1)
	}
	memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

	log.Print("Creating RabbitMQ client")
	rabbitmqClient := messagebroker.NewRabbitmqClient(appConfig.RabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	monitorError := monitor.Monitor(ctx, requester, memoryDatabase, messageBroker, appConfig)
	if monitorError != nil {
		log.Print(monitorError.Error())
		os.Exit(1)
	}
	log.Print("Execution finished")

	return nil
}
