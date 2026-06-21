package nslookup

import (
	"context"
	"net"
	"time"
)

// DNSLookup retrieves dns lookup information
// It provides DNS resolution functionality using a custom DNS server
type DNSLookup struct {
	DNSServer string // DNS server address (e.g., "8.8.8.8:53")
}

// getIP retrieves the IP address from the DNS lookup
// It performs a DNS lookup for the given domain using the configured DNS server
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - domain: Domain name to resolve
//
// Returns:
//   - string: Resolved IP address
//   - error: Error if DNS lookup fails
func (dnsLookup DNSLookup) Resolve(ctx context.Context, domain string) (string, error) {

	var ip string

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
		return ip, err
	} else {
		// Return the first IP address from the results
		ip = ips[0]
	}
	return ip, nil
}
