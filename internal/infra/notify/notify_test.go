//go:build integration_tests || unit_tests || notify_tests || notify_unit_tests

package notify

import (
	"context"
	"errors"
	"testing"

	opentelemetry "github.com/a-castellano/go-services/infra/opentelemetry"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	envelope "github.com/a-castellano/go-types/types/envelope"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// RabbitmqMock fakes the broker client. It records the last payload handed to
// SendMessage so tests can assert what actually travels on the wire.
// The SentMessage capture (and the pointer receivers it requires) was added by
// an AI agent (Claude); the original mock was hand-written.
type RabbitmqMock struct {
	LaunchError bool
	SentMessage []byte
}

func (client *RabbitmqMock) SendMessage(ctx context.Context, queueName string, message []byte) error {
	if client.LaunchError {
		return errors.New("Error")
	}
	client.SentMessage = message
	return nil
}

func (client *RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	if client.LaunchError {
		errorsChan <- errors.New("Error")
	} else {
		okMessage := []byte("This is ok")
		messages <- okMessage
		errorsChan <- nil
	}
}

// TestNotify covers the happy path and the wire contract: the payload handed
// to the broker is not the raw message but a marshalled envelope whose body
// carries it. The envelope assertions were written by an AI agent (Claude);
// the original happy-path test was hand-written.
func TestNotify(t *testing.T) {

	ctx := context.Background()

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: &rabbitmock}

	brokerNotifier := BrokerNotifier{Broker: messageBroker}

	testMessage := "This is a test"

	notifyError := brokerNotifier.Notify(ctx, "testQueue", testMessage)

	if notifyError != nil {
		t.Errorf("TestNotify should not fail")
	}

	sentEnvelope, unmarshalError := envelope.Unmarshal(rabbitmock.SentMessage)

	if unmarshalError != nil {
		t.Fatalf("TestNotify failed, sent payload should be a marshalled envelope, unmarshal error was '%s'.", unmarshalError.Error())
	}

	if string(sentEnvelope.Body) != testMessage {
		t.Errorf("TestNotify failed, envelope body should be '%s', it was '%s'.", testMessage, string(sentEnvelope.Body))
	}
}

// TestNotifyBrokerError covers the broker failure path: the error must
// propagate up to the caller without panicking.
// This test was written by an AI agent (Claude).
func TestNotifyBrokerError(t *testing.T) {

	ctx := context.Background()

	rabbitmock := RabbitmqMock{LaunchError: true}
	messageBroker := messagebroker.MessageBroker{Client: &rabbitmock}

	brokerNotifier := BrokerNotifier{Broker: messageBroker}

	notifyError := brokerNotifier.Notify(ctx, "testQueue", "This is a test")

	if notifyError == nil {
		t.Errorf("TestNotifyBrokerError should fail when the broker fails to send the message")
	}
}

// TestNotifyInjectsTraceContext checks the producer side of the distributed
// trace: Notify must open a PRODUCER span named after the queue and inject
// that same span's identity into the envelope carrier, so the consumer's
// Extract continues the trace as a child of the publish. It registers the
// global tracer provider and propagator itself so it does not depend on test
// order.
// This test was written by an AI agent (Claude).
func TestNotifyInjectsTraceContext(t *testing.T) {

	spanRecorder := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder)))
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	ctx := context.Background()

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: &rabbitmock}

	brokerNotifier := BrokerNotifier{Broker: messageBroker}

	notifyError := brokerNotifier.Notify(ctx, "testQueue", "This is a test")

	if notifyError != nil {
		t.Fatalf("TestNotifyInjectsTraceContext should not fail sending, error was '%s'.", notifyError.Error())
	}

	endedSpans := spanRecorder.Ended()
	if len(endedSpans) != 1 {
		t.Fatalf("TestNotifyInjectsTraceContext failed, Notify should end exactly one span, it ended %d.", len(endedSpans))
	}

	producerSpan := endedSpans[0]

	if producerSpan.SpanKind() != oteltrace.SpanKindProducer {
		t.Errorf("TestNotifyInjectsTraceContext failed, span kind should be producer, it was '%s'.", producerSpan.SpanKind().String())
	}

	if producerSpan.Name() != "send testQueue" {
		t.Errorf("TestNotifyInjectsTraceContext failed, span name should be 'send testQueue', it was '%s'.", producerSpan.Name())
	}

	sentEnvelope, unmarshalError := envelope.Unmarshal(rabbitmock.SentMessage)
	if unmarshalError != nil {
		t.Fatalf("TestNotifyInjectsTraceContext failed, sent payload should be a marshalled envelope, unmarshal error was '%s'.", unmarshalError.Error())
	}

	// Mirror the consumer side: extract on a brand new context, as the
	// consumer process would do after receiving the payload.
	extractedContext := opentelemetry.Extract(context.Background(), sentEnvelope)
	extractedSpanContext := oteltrace.SpanContextFromContext(extractedContext)

	if extractedSpanContext.TraceID() != producerSpan.SpanContext().TraceID() {
		t.Errorf("TestNotifyInjectsTraceContext failed, extracted TraceID should match the producer span's TraceID")
	}

	if extractedSpanContext.SpanID() != producerSpan.SpanContext().SpanID() {
		t.Errorf("TestNotifyInjectsTraceContext failed, extracted SpanID should match the producer span's SpanID, the producer span must be the remote parent")
	}
}
