package notify

import (
	"context"

	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
)

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
func Notify(ctx context.Context, broker messagebroker.MessageBroker, queueName string, message []byte) error {

	notifyError := broker.SendMessage(ctx, queueName, message)

	return notifyError

}
