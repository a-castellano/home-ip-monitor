package app

import (
	"context"
	"fmt"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

type Settings struct {
	ISPName     string
	DomainName  string
	NotifyQueue string
	UpdateQueue string
}

type Monitor struct {
	provider domain.IPInfoProvider
	resolver domain.DNSResolver
	store    domain.IPStore
	notifier domain.Notifier
	settings Settings
}

func NewMonitor(provider domain.IPInfoProvider, resolver domain.DNSResolver, storage domain.IPStore, notifier domain.Notifier, settings Settings) Monitor {
	return Monitor{provider: provider, resolver: resolver, store: storage, notifier: notifier, settings: settings}
}

func (monitor Monitor) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Starting monitor", "settings", monitor.settings, "operation", "Run")

	log.DebugContext(ctx, "Retrieving ipinfo data", "operation", "Run")

	ipinfo, getIPInfoErr := monitor.provider.GetIPInfo(ctx)

	if getIPInfoErr != nil {
		log.ErrorContext(ctx, "Error retrieving ipinfo data", "error", getIPInfoErr, "operation", "Run")
		return getIPInfoErr
	}

	log.DebugContext(ctx, "Validating that ipinfo provider is the esxpected provider", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Run")

	if !ipinfo.BelongsToISP(monitor.settings.ISPName) {
		log.DebugContext(ctx, "Current provider is not the expected provider, notifying only", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Run")

		notifyMessage := []byte(fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipinfo.IP, ipinfo.OrgName, monitor.settings.ISPName))

		notifyError := monitor.notifier.Notify(ctx, monitor.settings.NotifyQueue, notifyMessage)

		if notifyError != nil {

		}

	} else {
		log.DebugContext(ctx, "Current provider is the expected provider, cheking if IP has changed", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Run")

	}

	return nil
}

//
//func Monitor(ctx context.Context, ipInfoRequester ipinfo.Requester, nsLookup nslookup.NSLookup, memoryDatabase memorydatabase.MemoryDatabase, messageBroker messagebroker.MessageBroker, appConfig *config.Config) error {
//
//	log.Print("Retrieving IP info")
//	// Fetch current public IP and ISP information
//	ipInfo, ipInfoError := ipinfo.RetrieveIPInfoFromResponse(ipInfoRequester)
//
//	if ipInfoError != nil {
//		return ipInfoError
//	}
//
//	// Validate that the current ISP matches the expected provider
//	if ipInfo.OrgName != appConfig.ISPName {
//		notifyMessage := fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipInfo.IP, ipInfo.OrgName, appConfig.ISPName)
//
//		log.Print(notifyMessage)
//
//		// Send notification about ISP change without updating storage
//		encodedNotifyMessage := []byte(notifyMessage)
//		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
//		// end function, if notifyError is nil, final error is also nil as expected
//		return notifyError
//	}
//
//	log.Print("Checking IP info in storage.")
//	// Check if the current IP differs from the stored IP
//	requireUpdate, storageError := storage.CheckDatabase(ctx, ipInfo.IP, memoryDatabase)
//	if storageError != nil {
//		return storageError
//	}
//
//	log.Printf("IP update required: %v", requireUpdate)
//
//	// If no local change detected, verify against DNS resolution
//	if !requireUpdate {
//		log.Printf("Checking if remote IP matches with stored IP.")
//		// Perform DNS lookup to get the IP associated with the domain
//		remoteIP, nsLookupError := nslookup.GetIP(ctx, nsLookup, appConfig.DomainName)
//		if nsLookupError != nil {
//			return nsLookupError
//		}
//
//		// Update required if DNS IP differs from current IP
//		requireUpdate = remoteIP != ipInfo.IP
//
//	}
//
//	// Process IP change if update is required
//	if requireUpdate {
//
//		notifyMessage := fmt.Sprintf("Home IP has changed to %s.", ipInfo.IP)
//		log.Print(notifyMessage)
//
//		// Send notification message
//		encodedNotifyMessage := []byte(notifyMessage)
//		encodedIP := []byte(ipInfo.IP)
//		notifyError := notify.Notify(messageBroker, appConfig.NotifyQueue, encodedNotifyMessage)
//		if notifyError != nil {
//			return notifyError
//		}
//		// Send IP update message for DNS/record updates
//		notifyError = notify.Notify(messageBroker, appConfig.UpdateQueue, encodedIP)
//		if notifyError != nil {
//			return notifyError
//		}
//
//		// Update IP only after successful notification
//		log.Print("Updating IP in storage")
//		updateError := storage.UpdateIP(ctx, ipInfo.IP, memoryDatabase)
//
//		if updateError != nil {
//			return updateError
//		}
//	}
//
//	return nil
//}
