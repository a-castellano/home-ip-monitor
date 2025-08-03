//go:build integration_tests || unit_tests || ipinfo_tests || ipinfo_unit_tests

package nslookup

import (
	"context"
	"errors"
	"testing"
)

type MockResolver struct {
	Response      string
	ResponseError error
}

func (mock *MockResolver) getIP(ctx context.Context, domain string) (string, error) {
	return mock.Response, mock.ResponseError
}

func TestMockResolverCorrectIP(t *testing.T) {

	dnsLookup := MockResolver{Response: "1.1.1.1", ResponseError: nil}

	ctx := context.Background()

	domain := "test.windmaker.net"
	expectedIP := "1.1.1.1"

	ip, err := GetIP(ctx, &dnsLookup, domain)
	if err != nil {
		t.Errorf("GetIP should not fail resolving test.windmaker.net using MockResolver: %v", err)
	} else {
		if ip != expectedIP {
			t.Errorf("GetIP returned %s, expected %s", ip, expectedIP)
		}
	}
}

func TestMockResolverWithError(t *testing.T) {

	dnsLookup := MockResolver{Response: "1.1.1.1", ResponseError: errors.New("FAIL")}

	ctx := context.Background()

	domain := "test.windmaker.net"

	_, err := GetIP(ctx, &dnsLookup, domain)
	if err == nil {
		t.Errorf("GetIP should fail resolving test.windmaker.net using MockResolver: %v", err)
	}
}
