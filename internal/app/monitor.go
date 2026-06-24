package app

import (
	"context"
	"fmt"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

// Settings holds the four business values the use case needs. It is a plain
// value object so the application layer never sees infra wiring (Redis/RabbitMQ
// configs live in infra/config and are mapped to this in the composition root).
type Settings struct {
	ISPName     string
	DomainName  string
	NotifyQueue string
	UpdateQueue string
}

// Monitor is the application use case. All its dependencies are domain ports
// (interfaces), so it has zero knowledge of HTTP, Redis or RabbitMQ.
type Monitor struct {
	provider domain.IPInfoProvider
	resolver domain.DNSResolver
	store    domain.IPStore
	notifier domain.Notifier
	settings Settings
}

// NewMonitor builds a Monitor from its injected ports and settings. Since every
// field is unexported, this constructor is the only way to create a Monitor.
func NewMonitor(provider domain.IPInfoProvider, resolver domain.DNSResolver, storage domain.IPStore, notifier domain.Notifier, settings Settings) Monitor {
	return Monitor{provider: provider, resolver: resolver, store: storage, notifier: notifier, settings: settings}
}

// Run executes the monitoring flow:
//
//	Rule 1: read the current public IP and confirm it belongs to the expected ISP.
//	        If it does not, notify (only) and stop without touching storage.
//	Rule 2: compare the current IP with the stored one. If there is no stored IP
//	        or it differs, an update is required.
//	Rule 3: if it looks unchanged locally, cross-check against the domain's DNS
//	        record; a mismatch there also requires an update.
//	Rule 4: on update, notify both queues and only then persist the new IP, so a
//	        failed notification never leaves storage ahead of the notifications.
func (monitor Monitor) Run(ctx context.Context) error {

	log := logger.FromContext(ctx).With("operation", "Monitor.Run")
	log.DebugContext(ctx, "Starting monitor", "settings", monitor.settings)

	log.DebugContext(ctx, "Retrieving ipinfo data")

	// Rule 1: fetch the current public IP info.
	ipinfo, getIPInfoErr := monitor.provider.GetIPInfo(ctx)

	if getIPInfoErr != nil {
		log.ErrorContext(ctx, "Error retrieving ipinfo data", "error", getIPInfoErr)
		return getIPInfoErr
	}

	log.DebugContext(ctx, "Validating that ipinfo provider is the expected provider", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)

	// Rule 1: the IP must belong to the expected ISP. If not, notify and stop:
	// we do not update storage because this IP is not the home connection.
	if !ipinfo.BelongsToISP(monitor.settings.ISPName) {
		return monitor.notifyDifferentISP(ctx, ipinfo)
	}

	log.DebugContext(ctx, "Current provider is the expected provider, checking if IP has changed by retrieving the current stored IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)

	// Rules 2 & 3: decide whether the stored IP needs updating.
	updateIP, updateRequiredErr := monitor.updateRequired(ctx, ipinfo)
	if updateRequiredErr != nil {
		return updateRequiredErr
	}

	// Rule 4: notify both queues, then persist (notify-before-persist order).
	if updateIP {
		return monitor.applyUpdate(ctx, ipinfo)
	}

	return nil
}

// notifyDifferentISP handles Rule 1's notify-only path: the current IP does not
// belong to the expected ISP, so we notify and stop without touching storage.
func (monitor Monitor) notifyDifferentISP(ctx context.Context, ipinfo domain.IPInfo) error {

	log := logger.FromContext(ctx).With("operation", "Monitor.Run")
	log.DebugContext(ctx, "Current provider is not the expected provider, notifying only", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)

	notifyMessage := []byte(fmt.Sprintf("Read IP %s belongs to %s ISP, it seems that home is not using main ISP %s.", ipinfo.IP, ipinfo.OrgName, monitor.settings.ISPName))

	notifyError := monitor.notifier.Notify(ctx, monitor.settings.NotifyQueue, notifyMessage)

	if notifyError != nil {
		log.ErrorContext(ctx, "Error notifying about ISP change", "error", notifyError)
		return notifyError
	}

	return nil
}

// updateRequired implements Rules 2 & 3: it compares the current IP against the
// stored one and, when they look unchanged locally, cross-checks the domain's
// live DNS record. It returns whether an update is required (and any read error).
func (monitor Monitor) updateRequired(ctx context.Context, ipinfo domain.IPInfo) (bool, error) {

	log := logger.FromContext(ctx).With("operation", "Monitor.Run")

	// Rule 2: compare the current IP against the stored one.
	storedIP, ipFound, retrieveIPErr := monitor.store.StoredIP(ctx)

	if retrieveIPErr != nil {
		log.ErrorContext(ctx, "Error retrieving current stored IP from store", "error", retrieveIPErr)
		return false, retrieveIPErr
	}

	if !ipFound {
		log.DebugContext(ctx, "There is no stored IP, update with current value", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)
		return true, nil
	}

	log.DebugContext(ctx, "There is already an IP stored, compare with current IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "storedIP", storedIP)
	if storedIP != ipinfo.IP {
		log.DebugContext(ctx, "IPs differ, stored IP must be updated", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "storedIP", storedIP)
		return true, nil
	}
	log.DebugContext(ctx, "IPs are the same, stored IP will not be updated", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "storedIP", storedIP)

	// Rule 3: storage says it is unchanged, but cross-check against the
	// domain's live DNS record in case storage drifted from reality.
	log.DebugContext(ctx, "Stored IP matches, cross-checking against domain DNS resolution", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "domain", monitor.settings.DomainName)

	retrievedIPFromDNS, dnsRetrievalErr := monitor.resolver.Resolve(ctx, monitor.settings.DomainName)

	if dnsRetrievalErr != nil {
		log.ErrorContext(ctx, "Error resolving domain IP", "error", dnsRetrievalErr, "domain", monitor.settings.DomainName)
		return false, dnsRetrievalErr
	}

	if retrievedIPFromDNS != ipinfo.IP {
		log.DebugContext(ctx, "IP from domain DNS resolution differs from ipinfo IP, updating IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "domain", monitor.settings.DomainName, "retrievedIPFromDNS", retrievedIPFromDNS)
		return true, nil
	}

	log.DebugContext(ctx, "IP from domain DNS resolution matches ipinfo IP, update is not required", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP, "domain", monitor.settings.DomainName, "retrievedIPFromDNS", retrievedIPFromDNS)
	return false, nil
}

// applyUpdate implements Rule 4: it notifies both queues and only then persists
// the new IP, so a failed notification never leaves storage ahead of the
// notifications.
func (monitor Monitor) applyUpdate(ctx context.Context, ipinfo domain.IPInfo) error {

	log := logger.FromContext(ctx).With("operation", "Monitor.Run")

	log.DebugContext(ctx, "Notifying about IP change", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)
	notifyChangeMessage := fmt.Sprintf("Home IP has changed to %s.", ipinfo.IP)

	// Send notification message
	encodedNotifyChangeMessage := []byte(notifyChangeMessage)
	encodedIP := []byte(ipinfo.IP)

	notifyChangeError := monitor.notifier.Notify(ctx, monitor.settings.NotifyQueue, encodedNotifyChangeMessage)

	if notifyChangeError != nil {
		log.ErrorContext(ctx, "Error notifying about Home IP change", "error", notifyChangeError)
		return notifyChangeError
	}

	log.DebugContext(ctx, "Notifying about IP change in DNS update queue", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)

	notifyDNSError := monitor.notifier.Notify(ctx, monitor.settings.UpdateQueue, encodedIP)
	if notifyDNSError != nil {
		log.ErrorContext(ctx, "Error notifying DNS queue with IP to change", "error", notifyDNSError)
		return notifyDNSError
	}

	log.DebugContext(ctx, "Updating stored IP", "currentProvider", ipinfo.OrgName, "expectedProvider", monitor.settings.ISPName, "currentIP", ipinfo.IP)

	updateIPError := monitor.store.SaveIP(ctx, ipinfo.IP)
	if updateIPError != nil {
		log.ErrorContext(ctx, "Error updating retrieved IP in store", "error", updateIPError)
		return updateIPError
	}

	return nil
}
