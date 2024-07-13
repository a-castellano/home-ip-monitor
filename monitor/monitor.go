package monitor

import (
	"context"
	"fmt"
	"log"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	ipinfo "github.com/a-castellano/home-ip-monitor/ipinfo"
	notify "github.com/a-castellano/home-ip-monitor/notify"
	storage "github.com/a-castellano/home-ip-monitor/storage"
)

func Monitor(ctx context.Context, ipInfoRequester ipinfo.Requester, memoryDatabase memorydatabase.MemoryDatabase, messageBroker messagebroker.MessageBroker, appConfig *config.Config) error {

	log.Print("Retrieving IP info")
	ipInfo, ipInfoError := ipinfo.RetireveIPInfoFromResponse(ipInfoRequester)

	if ipInfoError != nil {
		return ipInfoError
	}

	if ipInfo.OrgName != appConfig.ISPName {
		notifyMessage := fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipInfo.IP, ipInfo.OrgName, appConfig.ISPName)

		log.Print(notifyMessage)

		encodedNotifyMessage := []byte(notifyMessage)
		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
		// end function, if notifyError is nill, final error is also nil asexpected
		return notifyError
	}

	log.Print("Checking IP info in storage.")
	requireUpdate, storageError := storage.CheckDatabase(ctx, ipInfo.IP, memoryDatabase)
	if storageError != nil {
		return storageError
	}

	log.Printf("IP update required: %v", requireUpdate)

	if requireUpdate {

		notifyMessage := fmt.Sprintf("Home IP has changed to %s.", ipInfo.IP)
		log.Print(notifyMessage)

		encodedNotifyMessage := []byte(notifyMessage)
		encodedIP := []byte(ipInfo.IP)
		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
		if notifyError != nil {
			return notifyError
		}
		notifyError = notify.Notify(messageBroker, appConfig.UpdateQueue, encodedIP)
		if notifyError != nil {
			return notifyError
		}

		// Update IP only after notify
		log.Print("Updating IP in storage")
		updateError := storage.UpdateIP(ctx, ipInfo.IP, memoryDatabase)

		if updateError != nil {
			return updateError
		}
	}

	return nil
}
