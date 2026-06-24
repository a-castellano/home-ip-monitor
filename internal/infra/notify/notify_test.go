//go:build integration_tests || unit_tests || notify_tests || notify_unit_tests

package notify

import (
	"context"
	"errors"
	messagebroker "github.com/a-castellano/go-services/services/messagebroker"
	"testing"
)

type RabbitmqMock struct {
	LaunchError bool
}

func (client RabbitmqMock) SendMessage(ctx context.Context, queueName string, message []byte) error {
	if client.LaunchError {
		return errors.New("Error")
	}
	return nil
}

func (client RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	if client.LaunchError {
		errorsChan <- errors.New("Error")
	} else {
		okMessage := []byte("This is ok")
		messages <- okMessage
		errorsChan <- nil
	}
}

// TestNotify covers BrokerNotifier.Notify. Notify is a thin pass-through to the
// broker's SendMessage with no logic of its own, so this happy-path test is
// enough: if the notification fails, the error simply propagates up from the
// broker through Notify to the use case and ultimately to main. There is no
// extra behaviour here worth a dedicated error-path test.
func TestNotify(t *testing.T) {

	ctx := context.Background()

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	brokerNotifier := BrokerNotifier{Broker: messageBroker}

	testMessage := []byte("This is a test")

	notifyError := brokerNotifier.Notify(ctx, "testQueue", testMessage)

	if notifyError != nil {
		t.Errorf("TestNotify should not fail")
	}
}
