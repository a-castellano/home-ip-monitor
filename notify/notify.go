package notify

import (
	messagebroker "github.com/a-castellano/go-services/messagebroker"
)

func Notify(broker messagebroker.MessageBroker, queueName string, message []byte) error {

	notifyError := broker.SendMessage(queueName, message)

	return notifyError

}
