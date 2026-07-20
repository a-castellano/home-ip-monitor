package nslookup

import (
	"context"
	"net"
	"time"

	logger "github.com/a-castellano/go-services/infra/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/a-castellano/home-ip-monitor/internal/infra/nslookup"

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
	ctx, span := otel.Tracer(tracerName).Start(ctx, "Resolve",
		trace.WithAttributes(
			attribute.String("operation", "Resolve"),
			attribute.String("dns.question.name", domain),
			attribute.String("dns.server", dnsLookup.DNSServer),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "Resolve")
	var ip string

	log.DebugContext(ctx, "creating dialer and resolver")
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
		errorString := "Error during domain nslookup"

		// RecordError stays here: the error is born in this span, no child
		// records it.
		span.RecordError(err)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "domain", domain, "error", err.Error())
		return ip, err
	}
	// Return the first IP address from the results
	ip = ips[0]
	span.SetAttributes(
		attribute.String("dns.resolved.ip", ip),
	)

	log.InfoContext(ctx, "domain ip retrieved", "domain", domain, "ip", ip)

	return ip, nil
}
