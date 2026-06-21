package nslookup

import (
	"context"
	logger "github.com/a-castellano/go-services/infra/logger"
	"net"
	"time"
)

// DNSLookup retrieves dns lookup information
// It provides DNS resolution functionality using a custom DNS server
type DNSLookup struct {
	DNSServer string // DNS server address (e.g., "8.8.8.8:53")
}

// Resolve resolves the given domain to an IP address using the configured DNS
// server. It implements domain.DNSResolver.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - domain: Domain name to resolve
//
// Returns:
//   - string: Resolved IP address (the first result)
//   - error: Error if DNS lookup fails
func (dnsLookup DNSLookup) Resolve(ctx context.Context, domain string) (string, error) {

	log := logger.FromContext(ctx)
	var ip string

	log.DebugContext(ctx, "Creating dialer and resolver", "operation", "Resolve")
	// Create dialer with timeout for DNS connections
	dialer := &net.Dialer{
		Timeout: time.Second * 5,
	}

	// Create custom resolver using the configured DNS server
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, dnsLookup.DNSServer)
		},
	}

	// Perform DNS lookup for the domain
	ips, err := resolver.LookupHost(ctx, domain)
	if err != nil {
		log.ErrorContext(ctx, "Error during domain nslookup", "domain", domain, "error", err.Error(), "operation", "Resolve")
		return ip, err
	} else {
		// Return the first IP address from the results
		ip = ips[0]
	}
	log.InfoContext(ctx, "domain ip retrived", "domain", domain, "ip", ip, "operation", "Resolve")

	return ip, nil
}
