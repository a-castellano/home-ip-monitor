package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	logger "github.com/a-castellano/go-services/infra/logger"
	rabbitmq "github.com/a-castellano/go-services/infra/rabbitmq"
	redis "github.com/a-castellano/go-services/infra/redis"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	slogconfig "github.com/a-castellano/go-types/slog"
	app "github.com/a-castellano/home-ip-monitor/internal/app"
	config "github.com/a-castellano/home-ip-monitor/internal/infra/config"
	ipinfodata "github.com/a-castellano/home-ip-monitor/internal/infra/ipinfodata"
	notify "github.com/a-castellano/home-ip-monitor/internal/infra/notify"
	nslookup "github.com/a-castellano/home-ip-monitor/internal/infra/nslookup"
	storage "github.com/a-castellano/home-ip-monitor/internal/infra/storage"
)

func main() {

	// First, initiate logger
	logConfig, err := slogconfig.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	appLogger := logger.NewLogger(logConfig)
	ctx := logger.WithLogger(context.Background(), appLogger)

	// Now from anywhere else in your program, you can use this:
	appLogger.DebugContext(ctx, "Loading config")

	appConfig, configErr := config.NewConfig()

	if configErr != nil {
		appLogger.ErrorContext(ctx, "Error loading app config", "error", configErr)
		os.Exit(1)
	}

	appLogger.InfoContext(ctx, "Initiating required services")
	appLogger.DebugContext(ctx, "Defining http client use by ipinfo package")

	httpClient := http.Client{
		Timeout: time.Second * 5,
	}

	appLogger.DebugContext(ctx, "Defining ipinfo requester")
	requester := ipinfodata.IPInfoRequester{HttpClient: &httpClient}

	appLogger.DebugContext(ctx, "Defining nslookup resolver")
	nsLookup := nslookup.DNSLookup{DNSServer: appConfig.DNSServer}

	appLogger.DebugContext(ctx, "Defining rabbitmq instance")
	rabbitmqClient := rabbitmq.NewRabbitmqClient(appConfig.RabbitmqConfig)
	appLogger.DebugContext(ctx, "Defining messagebroker instance")
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	appLogger.DebugContext(ctx, "Defining notifier instance")
	notifier := notify.BrokerNotifier{Broker: messageBroker}

	appLogger.DebugContext(ctx, "Defining redis instance")
	redisClient := redis.NewRedisClient(appConfig.RedisConfig)

	appLogger.DebugContext(ctx, "Initiating redis instance")
	if redisErr := redisClient.Initiate(ctx); redisErr != nil {
		appLogger.ErrorContext(ctx, "Error initiating redis instance", "error", redisErr)
		os.Exit(1)
	}

	appLogger.DebugContext(ctx, "Defining memorydatabase instance")
	memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

	appLogger.DebugContext(ctx, "Defining store instance")
	store := storage.Store{Database: memoryDatabase}

	monitorSettings := app.Settings{ISPName: appConfig.ISPName, DomainName: appConfig.DomainName, NotifyQueue: appConfig.NotifyQueue, UpdateQueue: appConfig.UpdateQueue}

	monitor := app.NewMonitor(requester, nsLookup, &store, &notifier, monitorSettings)
	// Start the monitoring process
	if monitorErr := monitor.Run(ctx); monitorErr != nil {
		appLogger.ErrorContext(ctx, "Error running monitor", "error", monitorErr)
		os.Exit(1)
	}
}
