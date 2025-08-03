//go:build integration_tests || nslookup_tests || nslookup_unit_tests

package nslookup

import (
	"context"
	"testing"
)

func TestGetIP(t *testing.T) {

	dnsLookup := DNSLookup{DNSServer: "1.1.1.1:53"}

	ctx := context.Background()

	domain := "test.windmaker.net"
	expectedIP := "213.32.122.25"

	ip, err := GetIP(ctx, &dnsLookup, domain)
	if err != nil {
		t.Errorf("GetIP should not fail resolving test.windmaker.net: %v", err)
	} else {
		if ip != expectedIP {
			t.Errorf("GetIP returned %s, expected %s", ip, expectedIP)
		}
	}
}
