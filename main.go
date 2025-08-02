package main

import (
	"context"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	"github.com/a-castellano/home-ip-monitor/ipinfo"
	"github.com/a-castellano/home-ip-monitor/monitor"
)

const serviceName = "home-ip-monitor"

func main() {

	// Configure logger to write to the syslog. You could do this in init(), too.
	logwriter, e := syslog.New(syslog.LOG_INFO, serviceName)
	if e == nil {
		log.SetOutput(logwriter)
		// Remove timestamp
		log.SetFlags(0)
	}

	// Now from anywhere else in your program, you can use this:
	log.Print("Loading config")

	appConfig, configErr := config.NewConfig()

	if configErr != nil {
		log.Print(configErr.Error())
		os.Exit(1)
	}

	log.Print("Defining http client used by ipInfo package")

	httpClient := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	ctx := context.Background()

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
}
