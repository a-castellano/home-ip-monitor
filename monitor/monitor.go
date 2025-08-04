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
	nslookup "github.com/a-castellano/home-ip-monitor/nslookup"
	storage "github.com/a-castellano/home-ip-monitor/storage"
)

// Monitor performs the core IP monitoring logic:
// 1. Fetches current public IP from ipinfo.io
// 2. Validates ISP consistency with expected provider
// 3. Compares with previously stored IP
// 4. Performs DNS verification if no local changes detected
// 5. Sends notifications and updates storage when changes occur
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - ipInfoRequester: Interface for fetching IP information
//   - nsLookup: Interface for DNS resolution
//   - memoryDatabase: Interface for data persistence
//   - messageBroker: Interface for message queuing
//   - appConfig: Application configuration
//
// Returns error if any step fails
func Monitor(ctx context.Context, ipInfoRequester ipinfo.Requester, nsLookup nslookup.NSLookup, memoryDatabase memorydatabase.MemoryDatabase, messageBroker messagebroker.MessageBroker, appConfig *config.Config) error {

	log.Print("Retrieving IP info")
	// Fetch current public IP and ISP information
	ipInfo, ipInfoError := ipinfo.RetrieveIPInfoFromResponse(ipInfoRequester)

	if ipInfoError != nil {
		return ipInfoError
	}

	// Validate that the current ISP matches the expected provider
	if ipInfo.OrgName != appConfig.ISPName {
		notifyMessage := fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipInfo.IP, ipInfo.OrgName, appConfig.ISPName)

		log.Print(notifyMessage)

		// Send notification about ISP change without updating storage
		encodedNotifyMessage := []byte(notifyMessage)
		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
		// end function, if notifyError is nil, final error is also nil as expected
		return notifyError
	}

	log.Print("Checking IP info in storage.")
	// Check if the current IP differs from the stored IP
	requireUpdate, storageError := storage.CheckDatabase(ctx, ipInfo.IP, memoryDatabase)
	if storageError != nil {
		return storageError
	}

	log.Printf("IP update required: %v", requireUpdate)

	// If no local change detected, verify against DNS resolution
	if !requireUpdate {
		log.Printf("Checking if remote IP matches with stored IP.")
		// Perform DNS lookup to get the IP associated with the domain
		remoteIP, nsLookupError := nslookup.GetIP(ctx, nsLookup, appConfig.DomainName)
		if nsLookupError != nil {
			return nsLookupError
		}

		// Update required if DNS IP differs from current IP
		requireUpdate = remoteIP != ipInfo.IP

	}

	// Process IP change if update is required
	if requireUpdate {

		notifyMessage := fmt.Sprintf("Home IP has changed to %s.", ipInfo.IP)
		log.Print(notifyMessage)

		// Send notification message
		encodedNotifyMessage := []byte(notifyMessage)
		encodedIP := []byte(ipInfo.IP)
		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
		if notifyError != nil {
			return notifyError
		}
		// Send IP update message for DNS/record updates
		notifyError = notify.Notify(messageBroker, appConfig.UpdateQueue, encodedIP)
		if notifyError != nil {
			return notifyError
		}

		// Update IP only after successful notification
		log.Print("Updating IP in storage")
		updateError := storage.UpdateIP(ctx, ipInfo.IP, memoryDatabase)

		if updateError != nil {
			return updateError
		}
	}

	return nil
}
