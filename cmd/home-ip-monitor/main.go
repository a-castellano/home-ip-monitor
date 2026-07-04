package main

import (
	"context"
	systemlog "log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	logger "github.com/a-castellano/go-services/infra/logger"
	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	rabbitmq "github.com/a-castellano/go-services/infra/rabbitmq"
	redis "github.com/a-castellano/go-services/infra/redis"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	otelconfig "github.com/a-castellano/go-types/types/opentelemetry"
	slogconfig "github.com/a-castellano/go-types/types/slog"
	app "github.com/a-castellano/home-ip-monitor/internal/app"
	config "github.com/a-castellano/home-ip-monitor/internal/infra/config"
	ipinfodata "github.com/a-castellano/home-ip-monitor/internal/infra/ipinfodata"
	notify "github.com/a-castellano/home-ip-monitor/internal/infra/notify"
	nslookup "github.com/a-castellano/home-ip-monitor/internal/infra/nslookup"
	storage "github.com/a-castellano/home-ip-monitor/internal/infra/storage"
)

func run(ctx context.Context) error {
	log := logger.FromContext(ctx).With("operation", "main.run")
	log.DebugContext(ctx, "Loading config")

	otelConfig, otelConfigErr := otelconfig.NewConfig()
	if otelConfigErr != nil {
		log.ErrorContext(ctx, "telemetry config has errors", "error", otelConfigErr)
		return otelConfigErr
	}

	shutdown, err := opentelemetry.SetupOpenTelemetry(ctx, otelConfig)
	if err != nil {
		// Telemetry failed to start; the app keeps running without it.
		log.ErrorContext(ctx, "telemetry setup failed", "error", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.ErrorContext(ctx, "telemetry shutdown failed", "error", err)
		}
	}()

	appConfig, configErr := config.NewConfig(ctx)

	if configErr != nil {
		log.ErrorContext(ctx, "Error loading app config", "error", configErr)
		return configErr
	}

	log.InfoContext(ctx, "Initiating required services")
	log.DebugContext(ctx, "Defining http client use by ipinfo package")

	httpClient := http.Client{
		Timeout:   time.Second * 5,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	log.DebugContext(ctx, "Defining ipinfo requester")
	requester := ipinfodata.IPInfoRequester{HttpClient: &httpClient}

	log.DebugContext(ctx, "Defining nslookup resolver")
	nsLookup := nslookup.DNSLookup{DNSServer: appConfig.DNSServer}

	log.DebugContext(ctx, "Defining rabbitmq instance")
	rabbitmqClient := rabbitmq.NewRabbitmqClient(appConfig.RabbitmqConfig)
	log.DebugContext(ctx, "Defining messagebroker instance")
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	log.DebugContext(ctx, "Defining notifier instance")
	notifier := notify.BrokerNotifier{Broker: messageBroker}

	log.DebugContext(ctx, "Defining redis instance")
	redisClient := redis.NewRedisClient(appConfig.RedisConfig)

	log.DebugContext(ctx, "Initiating redis instance")
	if redisErr := redisClient.Initiate(ctx); redisErr != nil {
		log.ErrorContext(ctx, "Error initiating redis instance", "error", redisErr)
		return redisErr
	}

	log.DebugContext(ctx, "Defining memorydatabase instance")
	memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

	log.DebugContext(ctx, "Defining store instance")
	store := storage.Store{Database: memoryDatabase}

	monitorSettings := app.Settings{ISPName: appConfig.ISPName, DomainName: appConfig.DomainName, NotifyQueue: appConfig.NotifyQueue, UpdateQueue: appConfig.UpdateQueue}

	monitor := app.NewMonitor(requester, nsLookup, &store, &notifier, monitorSettings)
	// Start the monitoring process
	if monitorErr := monitor.Run(ctx); monitorErr != nil {
		log.ErrorContext(ctx, "Error running monitor", "error", monitorErr)
		return monitorErr
	}

	return nil
}

func main() {

	// First, initiate logger
	logConfig, err := slogconfig.NewConfig()
	if err != nil {
		systemlog.Fatal(err)
	}

	appLogger := logger.NewLogger(logConfig)
	ctx := logger.WithLogger(context.Background(), appLogger)

	runError := run(ctx)

	if runError != nil {
		appLogger.ErrorContext(ctx, "error running home-ip-monitor", "error", runError)
		os.Exit(1)
	}

}
