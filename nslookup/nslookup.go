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
func (dnsLookup DNSLookup) GetIP(ctx context.Context, domain string) (string, error) {

	var ip string = ""

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

// NSLookup is an interface that defines the method to retrieve IP information
// This interface allows for easy testing by mocking DNS operations
type NSLookup interface {
	GetIP(context.Context, string) (string, error)
}

// GetIP retrieves the IP address using the provided NSLookup interface
// It's a wrapper function that delegates to the interface implementation
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - nsLookup: Interface for DNS operations
//   - domain: Domain name to resolve
//
// Returns:
//   - string: Resolved IP address
//   - error: Error if DNS lookup fails
func GetIP(ctx context.Context, nsLookup NSLookup, domain string) (string, error) {

	return nsLookup.GetIP(ctx, domain)

}
