//go:build integration_tests || unit_tests || app_tests || app_unit_tests

package app

import (
	"context"
	"errors"
	"testing"

	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

// Mock all domain interfaces

type IPInfoMock struct {
	IPInfoData domain.IPInfo
	Error      error
}

func (mock IPInfoMock) GetIPInfo(ctx context.Context) (domain.IPInfo, error) {
	return mock.IPInfoData, mock.Error
}

type DNSResolverMock struct {
	Result string
	Error  error
}

func (mock DNSResolverMock) Resolve(ctx context.Context, domain string) (string, error) {
	return mock.Result, mock.Error
}

type IPStoreMock struct {
	StoredIPValue string
	StoreFound    bool
	StoreError    error

	SaveError error
}

func (mock IPStoreMock) StoredIP(ctx context.Context) (string, bool, error) {
	return mock.StoredIPValue, mock.StoreFound, mock.StoreError
}

func (mock IPStoreMock) SaveIP(ctx context.Context, ip string) error {
	return mock.SaveError
}

type NotifierMock struct {
	Error error
}

func (mock NotifierMock) Notify(ctx context.Context, queue string, message []byte) error {
	return mock.Error
}

type ComplexNotifierMock struct {
	FailQueue string
	Error     error
}

func (mock ComplexNotifierMock) Notify(ctx context.Context, queue string, message []byte) error {
	if queue == mock.FailQueue {
		return mock.Error
	}
	return nil
}

func TestErrorOnGetIPInfo(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: errors.New("failed to fetch")}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.2.3.4", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Example", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorOnGetIPInfo should fail because mocked IPInfo returns an error")
	}

}

func TestErrorDifferentProviderNotifyError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.2.3.4", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: errors.New("Fail")}

	settings := Settings{ISPName: "Different", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorDifferentProviderNotifyError should fail because mocked notifier returns an error and ISP differs")
	}

}

func TestErrorSameProviderCheckStoredIPError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.2.3.4", StoreFound: true, StoreError: errors.New("Fail"), SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorDifferentProviderNotifyError should fail because mocked notifier returns an error and ISP differs")
	}

}

func TestErrorSameProviderStoredIPIsTheSame(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.1", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestErrorSameProviderStoredIPIsTheSame should not fail")
	}

}

func TestErrorSameProviderStoredIPIsDiffer(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.2", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestErrorSameProviderStoredIPIsDiffer should not fail")
	}

}

func TestErrorSameProviderStoredIPNotFound(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.2", StoreFound: false, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestErrorSameProviderStoredIPNotFound should not fail")
	}

}

func TestErrorSameProviderStoredIPIsTheSameDNSResolveFails(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: errors.New("Fail")}

	store := IPStoreMock{StoredIPValue: "1.1.1.1", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorSameProviderStoredIPIsTheSameDNSResolveFails should fail")
	}

}

func TestErrorSameProviderStoredIPIsTheSameResolverReturnsSame(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "1.1.1.1", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.1", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestErrorSameProviderStoredIPIsTheSameResolverReturnsSame should not fail")
	}

}

func TestErrorSameProviderStoredIPIsDifferNotifyFails(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.2", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := NotifierMock{Error: errors.New("Fail")}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorSameProviderStoredIPIsDiffer should fail, because notify IP change fails")
	}

}

func TestErrorSameProviderStoredIPIsDifferNotifyIPFails(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.2", StoreFound: true, StoreError: nil, SaveError: nil}

	notifier := ComplexNotifierMock{Error: errors.New("Fail"), FailQueue: "update"}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorSameProviderStoredIPIsDiffer should fail, because notify IP update fails")
	}

}

func TestErrorSameProviderStoredIPIsDifferStoreSaveFails(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		IPInfoMock{IPInfoData: ipinfoData, Error: nil}

	resolver := DNSResolverMock{Result: "any", Error: nil}

	store := IPStoreMock{StoredIPValue: "1.1.1.2", StoreFound: true, StoreError: nil, SaveError: errors.New("Fail")}

	notifier := NotifierMock{Error: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestErrorSameProviderStoredIPIsDiffer should fail, because store fails when IP is saved")
	}

}
