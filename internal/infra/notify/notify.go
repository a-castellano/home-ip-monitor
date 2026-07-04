package notify

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/a-castellano/home-ip-monitor/internal/infra/notify"

// BrokerNotifier is the messaging adapter. It wraps a
// messagebroker.MessageBroker and implements domain.Notifier.
type BrokerNotifier struct {
	Broker messagebroker.MessageBroker
}

// Notify sends message to the given queue through the message broker.
// It implements domain.Notifier.
//
// The message does not travel raw: Notify opens a dedicated PRODUCER span for
// the publish, injects that span's trace context into an envelope together
// with the message, and hands the marshalled envelope to the broker. This is
// the wire contract with consumers: they must unmarshal the envelope first,
// extract the trace context from its carrier to continue the distributed
// trace, and read the message from its body.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts, carrying the parent span
//   - queue: Name of the queue to send the message to
//   - message: Message content
//
// Returns:
//   - error: Error if enveloping the message or sending it fails
func (brokerNotifier *BrokerNotifier) Notify(ctx context.Context, queue string, message string) error {

	ctx, span := otel.Tracer(tracerName).Start(ctx, "send "+queue,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("operation", "Notify"),
			attribute.String("messaging.destination.name", queue),
			attribute.String("messaging.operation.type", "send"),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "Notify")
	log.DebugContext(ctx, "Notifying message to queue", "queue", queue, "message", message)

	encodedMessage := []byte(message)
	envelope := opentelemetry.Inject(ctx, encodedMessage)

	envelopedMessage, err := envelope.Marshal()
	if err != nil {
		errorString := "failed to envelope message"

		span.RecordError(err)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "queue", queue, "message", message, "error", err.Error())
		return err
	}
	notifyError := brokerNotifier.Broker.SendMessage(ctx, queue, envelopedMessage)
	if notifyError != nil {
		errorString := "broker failed to send message"

		span.RecordError(notifyError)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "queue", queue, "message", message, "error", notifyError.Error())

	}

	return notifyError

}
