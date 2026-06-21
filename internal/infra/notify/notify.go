package notify

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
)

type BrokerNotifier struct {
	broker messagebroker.MessageBroker
}

// Notify sends a message to the specified queue using the message broker
// It's a simple wrapper around the message broker's SendMessage method
//
// Parameters:
//   - broker: Message broker interface for sending messages
//   - queueName: Name of the queue to send the message to
//   - message: Message content as byte array
//
// Returns:
//   - error: Error if message sending fails
func (brokerNotifier *BrokerNotifier) Notify(ctx context.Context, queue string, message []byte) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Notifying message to queue", "queue", queue, "message", message, "operation", "Notify")

	notifyError := brokerNotifier.broker.SendMessage(ctx, queue, message)

	return notifyError

}
