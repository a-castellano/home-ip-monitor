package monitor

import (
	"context"
	"fmt"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	ipinfo "github.com/a-castellano/home-ip-monitor/ipinfo"
	notify "github.com/a-castellano/home-ip-monitor/notify"
	"github.com/a-castellano/home-ip-monitor/storage"
)

func Monitor(ctx context.Context, ipInfoRequester ipinfo.Requester, memoryDatabase memorydatabase.MemoryDatabase, messageBroker messagebroker.MessageBroker, appConfig config.Config) error {

	ipInfo, ipInfoError := ipinfo.RetireveIPInfoFromResponse(ipInfoRequester)

	if ipInfoError != nil {
		return ipInfoError
	}

	if ipInfo.OrgName != appConfig.ISPName {
		notifyMessage := []byte(fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipInfo.IP, ipInfo.OrgName, appConfig.ISPName))
		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, notifyMessage)
		if notifyError != nil {
			return notifyError
		}
	} else {
		requireUpdate, storageError := storage.CheckDatabase(ctx, ipInfo.IP, memoryDatabase)
		if storageError != nil {
			return storageError
		}
		if requireUpdate {

			notifyMessage := []byte(fmt.Sprintf("Home IP has changed to %s.", ipInfo.IP))
			encodedIP := []byte(ipInfo.IP)
			notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, notifyMessage)
			if notifyError != nil {
				return notifyError
			}
			notifyError = notify.Notify(messageBroker, appConfig.UpdateQueue, encodedIP)
			if notifyError != nil {
				return notifyError
			}

			// Update IP only after notify

			updateError := storage.UpdateIP(ctx, ipInfo.IP, memoryDatabase)

			if updateError != nil {
				return updateError
			}
		}
	}

	return nil
}
