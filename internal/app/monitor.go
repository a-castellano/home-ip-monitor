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

	var updateIP bool = false

	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Starting monitor", "settings", monitor.settings, "operation", "Monitor.Run")

	log.DebugContext(ctx, "Retrieving ipinfo data", "operation", "Monitor.Run")

	ipinfo, getIPInfoErr := monitor.provider.GetIPInfo(ctx)

	if getIPInfoErr != nil {
		log.ErrorContext(ctx, "Error retrieving ipinfo data", "error", getIPInfoErr, "operation", "Monitor.Run")
		return getIPInfoErr
	}

	log.DebugContext(ctx, "Validating that ipinfo provider is the esxpected provider", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")

	if !ipinfo.BelongsToISP(monitor.settings.ISPName) {
		log.DebugContext(ctx, "Current provider is not the expected provider, notifying only", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")

		notifyMessage := []byte(fmt.Sprintf("Readed IP %s belongs to %s ISP, it seems than home is not using main ISP %s.", ipinfo.IP, ipinfo.OrgName, monitor.settings.ISPName))

		notifyError := monitor.notifier.Notify(ctx, monitor.settings.NotifyQueue, notifyMessage)

		if notifyError != nil {

			log.ErrorContext(ctx, "Error notifying about ISP change", "error", notifyError, "operation", "Monitor.Run")
			return notifyError
		}

	} else {
		log.DebugContext(ctx, "Current provider is the expected provider, cheking if IP has changed retrieving current stored ip", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")

		storedIP, ipFound, retrieveIPErr := monitor.store.StoredIP(ctx)

		if retrieveIPErr != nil {
			log.ErrorContext(ctx, "Error retrieving current stored IP from store", "error", retrieveIPErr, "operation", "Monitor.Run")

			return retrieveIPErr
		}

		if ipFound {
			log.DebugContext(ctx, "There is already an IP stored, compare with current IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "currentIP", storedIP, "operation", "Monitor.Run")
			if storedIP != ipinfo.IP {
				log.DebugContext(ctx, "IP's differ, stored IP must be updated", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "updateIP", storedIP, "operation", "Monitor.Run")
				updateIP = true
			} else {
				log.DebugContext(ctx, "IP's are the same, stored IP will no be updated", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "storedIP", storedIP, "operation", "Monitor.Run")
			}
		} else {
			log.DebugContext(ctx, "There is not stored IP, update with current value", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")
			updateIP = true
		}

		if !updateIP {
			log.DebugContext(ctx, "For the time being, IP does not require to be updated, cheking resolver", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "domain", monitor.settings.DomainName, "operation", "Monitor.Run")

			retriedIPFromDNS, dnsRetirevalErr := monitor.resolver.Resolve(ctx, monitor.settings.DomainName)

			if dnsRetirevalErr != nil {
				log.ErrorContext(ctx, "Error resolving domain ip", "error", dnsRetirevalErr, "domain", monitor.settings.DomainName, "operation", "Monitor.Run")
				return dnsRetirevalErr
			}

			if retriedIPFromDNS != ipinfo.IP {
				log.DebugContext(ctx, "retrieved IP from domain DNS resolution differs from ipinfo IP, updafing IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "domain", monitor.settings.DomainName, "retriedIPFromDNS", retriedIPFromDNS, "operation", "Monitor.Run")
				updateIP = true
			} else {
				log.DebugContext(ctx, "retrieved IP from domain DNS resolution does no differ from ipinfo IP, update is not required", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "domain", monitor.settings.DomainName, "retriedIPFromDNS", retriedIPFromDNS, "operation", "Monitor.Run")
			}

		}

		if updateIP {

			log.DebugContext(ctx, "Notifying about IP change", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")
			notifyChangeMessage := fmt.Sprintf("Home IP has changed to %s.", ipinfo.IP)

			// Send notification message
			encodedNotifyChangeMessage := []byte(notifyChangeMessage)
			encodedIP := []byte(ipinfo.IP)

			notifyChangeError := monitor.notifier.Notify(ctx, monitor.settings.NotifyQueue, encodedNotifyChangeMessage)

			if notifyChangeError != nil {
				log.ErrorContext(ctx, "Error notifying about Home IP change", "error", notifyChangeError, "operation", "Monitor.Run")
				return notifyChangeError
			}

			log.DebugContext(ctx, "Notifying about IP change in DNS update queue", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")

			notifyDNSError := monitor.notifier.Notify(ctx, monitor.settings.UpdateQueue, encodedIP)
			if notifyDNSError != nil {
				log.ErrorContext(ctx, "Error notifying DNS queue with IP to change", "error", notifyDNSError, "operation", "Monitor.Run")
				return notifyDNSError
			}

			log.DebugContext(ctx, "Updating stored IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "curretIP", ipinfo.IP, "operation", "Monitor.Run")

			updateIPError := monitor.store.SaveIP(ctx, ipinfo.IP)
			if updateIPError != nil {
				log.ErrorContext(ctx, "Error updating retrived IP in store", "error", updateIPError, "operation", "Monitor.Run")
				return updateIPError
			}

		}

	}

	return nil
}
