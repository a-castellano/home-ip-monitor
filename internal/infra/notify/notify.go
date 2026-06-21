package notify

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
)

// BrokerNotifier is the messaging adapter. It wraps a
// messagebroker.MessageBroker and implements domain.Notifier.
type BrokerNotifier struct {
	broker messagebroker.MessageBroker
}

// Notify sends message to the given queue through the message broker.
// It implements domain.Notifier.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - queue: Name of the queue to send the message to
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
