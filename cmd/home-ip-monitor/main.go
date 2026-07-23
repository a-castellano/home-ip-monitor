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
	log.DebugContext(ctx, "loading config")

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
		shutdownCtx, cancelShutdown := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancelShutdown()
		if err := shutdown(shutdownCtx); err != nil {
			log.ErrorContext(shutdownCtx, "telemetry shutdown failed", "error", err)
		}
	}()

	appConfig, configErr := config.NewConfig(ctx)

	if configErr != nil {
		log.ErrorContext(ctx, "error loading app config", "error", configErr)
		return configErr
	}

	log.InfoContext(ctx, "initiating required services")
	log.DebugContext(ctx, "defining HTTP client used by ipinfo package")

	httpClient := http.Client{
		Timeout:   time.Second * 5,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	log.DebugContext(ctx, "defining ipinfo requester")
	requester := ipinfodata.IPInfoRequester{HttpClient: &httpClient}

	log.DebugContext(ctx, "defining nslookup resolver")
	nsLookup := nslookup.DNSLookup{DNSServer: appConfig.DNSServer}

	log.DebugContext(ctx, "defining RabbitMQ instance")
	rabbitmqClient := rabbitmq.NewRabbitmqClient(appConfig.RabbitmqConfig)
	log.DebugContext(ctx, "defining messagebroker instance")
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	log.DebugContext(ctx, "defining notifier instance")
	notifier := notify.BrokerNotifier{Broker: messageBroker}

	log.DebugContext(ctx, "defining Redis instance")
	redisClient := redis.NewRedisClient(appConfig.RedisConfig)

	log.DebugContext(ctx, "initiating Redis instance")
	if redisErr := redisClient.Initiate(ctx); redisErr != nil {
		log.ErrorContext(ctx, "error initiating Redis instance", "error", redisErr)
		return redisErr
	}

	log.DebugContext(ctx, "defining memorydatabase instance")
	memoryDatabase := memorydatabase.NewMemoryDatabase(&redisClient)

	log.DebugContext(ctx, "defining store instance")
	store := storage.Store{Database: memoryDatabase}

	monitorSettings := app.Settings{ISPName: appConfig.ISPName, DomainName: appConfig.DomainName, NotifyQueue: appConfig.NotifyQueue, UpdateQueue: appConfig.UpdateQueue}

	monitor := app.NewMonitor(ctx, requester, nsLookup, &store, &notifier, monitorSettings)
	// Start the monitoring process
	if monitorErr := monitor.Run(ctx); monitorErr != nil {
		log.ErrorContext(ctx, "error running monitor", "error", monitorErr)
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

	appLogger := logger.NewLogger(logConfig, opentelemetry.NewSlogHandler(logConfig.AppName))
	ctx := logger.WithLogger(context.Background(), appLogger)

	runError := run(ctx)

	if runError != nil {
		appLogger.ErrorContext(ctx, "error running home-ip-monitor", "error", runError)
		os.Exit(1)
	}

}
