package main

import (
	"context"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	"github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	"github.com/a-castellano/home-ip-monitor/ipinfo"
	"github.com/a-castellano/home-ip-monitor/monitor"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func printMessage(ctx context.Context, message string) {
	// New Span
	_, span := otel.Tracer(serviceName).Start(ctx, message)
	defer span.End()
}

func initTracer() (*sdktrace.TracerProvider, error) {
	// Configurar el exportador OTLP para enviar trazas a un servidor usando gRPC
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(collectorURL),
		otlptracegrpc.WithInsecure(), // Quitar esta línea si usas TLS
	)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, err
	}

	// Crear el TracerProvider con el exportador y recursos adicionales
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// Establecer el TracerProvider global
	otel.SetTracerProvider(tp)
	return tp, nil
}

var (
	collectorURL = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
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
	log.Print("Initiating telemetry")

	tp, err := initTracer()
	if err != nil {
		log.Printf("Error starting spentelemetry tracer: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error closing opentelemetry tracer: %v", err)
			os.Exit(1)
		}
	}()

	// Crete context
	tracerContext, span := otel.Tracer(serviceName).Start(context.Background(), "main")
	defer span.End()

	printMessage(tracerContext, "Initiating services")

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
	rabbitmqClient := messagebroker.NewRabbimqClient(appConfig.RabbitmqConfig)
	messageBroker := messagebroker.MessageBroker{Client: rabbitmqClient}

	printMessage(tracerContext, "RunMonitor")
	monitorError := monitor.Monitor(ctx, requester, memoryDatabase, messageBroker, appConfig)
	if monitorError != nil {
		log.Print(monitorError.Error())
		os.Exit(1)
	}
	printMessage(tracerContext, "Execution finished")
	log.Print("Execution finished")
}
