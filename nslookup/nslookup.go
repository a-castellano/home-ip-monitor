package nslookup

import (
	"context"
	"net"
	"time"
)

// DNSLookup retrieves dns lookup information
type DNSLookup struct {
	DNSServer string
}

// getIP retrieves the IP address from the DNS lookup
func (dnsLookup DNSLookup) GetIP(ctx context.Context, domain string) (string, error) {

	var ip string = ""

	dialer := &net.Dialer{
		Timeout: time.Second * 5,
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, dnsLookup.DNSServer)
		},
	}

	ips, err := resolver.LookupHost(ctx, domain)
	if err != nil {
		return ip, err
	} else {
		ip = ips[0]
	}
	return ip, nil
}

// NSLookup is an interface that defines the method to retrieve IP information
type NSLookup interface {
	GetIP(context.Context, string) (string, error)
}

// GetIP retrieves the IP address using the provided NSLookup interface
func GetIP(ctx context.Context, nsLookup NSLookup, domain string) (string, error) {

	return nsLookup.GetIP(ctx, domain)

}
