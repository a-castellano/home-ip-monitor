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

func TestNotify(t *testing.T) {

	ctx := context.Background()

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	brokerNotifier := BrokerNotifier{broker: messageBroker}

	testMessage := []byte("This is a test")

	notifyError := brokerNotifier.Notify(ctx, "testQeue", testMessage)

	if notifyError != nil {
		t.Errorf("TestNotify should not fail")
	}
}
