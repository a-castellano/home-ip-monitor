//go:build integration_tests || unit_tests || app_tests || app_unit_tests

package app

import (
	"context"
	"errors"
	"testing"

	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

// The mocks below are hand-written fakes for the four domain ports. They return
// fixed values configured per test; they do not track calls, so the tests assert
// only on the error returned by Run (matching the repo's existing test style).

// ipInfoMock fakes domain.IPInfoProvider.
type ipInfoMock struct {
	ipInfoData domain.IPInfo
	err        error
}

func (mock ipInfoMock) GetIPInfo(ctx context.Context) (domain.IPInfo, error) {
	return mock.ipInfoData, mock.err
}

// dnsResolverMock fakes domain.DNSResolver.
type dnsResolverMock struct {
	result string
	err    error
}

func (mock dnsResolverMock) Resolve(ctx context.Context, domain string) (string, error) {
	return mock.result, mock.err
}

// ipStoreMock fakes domain.IPStore.
type ipStoreMock struct {
	storedIPValue string
	storeFound    bool
	storeError    error

	saveError error
}

func (mock ipStoreMock) StoredIP(ctx context.Context) (string, bool, error) {
	return mock.storedIPValue, mock.storeFound, mock.storeError
}

func (mock ipStoreMock) SaveIP(ctx context.Context, ip string) error {
	return mock.saveError
}

// notifierMock fakes domain.Notifier, returning the same result for every queue.
type notifierMock struct {
	err error
}

func (mock notifierMock) Notify(ctx context.Context, queue string, message []byte) error {
	return mock.err
}

// complexNotifierMock fakes domain.Notifier but fails only for a specific queue,
// so a test can exercise the second (UpdateQueue) notification independently.
type complexNotifierMock struct {
	failQueue string
	err       error
}

func (mock complexNotifierMock) Notify(ctx context.Context, queue string, message []byte) error {
	if queue == mock.failQueue {
		return mock.err
	}
	return nil
}

// Rule 1: when the provider itself fails, Run must propagate that error and do
// nothing else.
func TestGetIPInfoError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: errors.New("failed to fetch")}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.2.3.4", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Example", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestGetIPInfoError should fail because mocked IPInfo returns an error")
	}

}

// Rule 1: the IP belongs to a different ISP, so Run takes the notify-only path.
// Here that single notification fails, so Run must return its error.
func TestDifferentISPNotifyError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.2.3.4", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: errors.New("Fail")}

	settings := Settings{ISPName: "Different", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestDifferentISPNotifyError should fail because mocked notifier returns an error and ISP differs")
	}

}

// Rule 2: the ISP matches, but reading the stored IP fails. Run must propagate
// the store read error.
func TestStoredIPReadError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.2.3.4", storeFound: true, storeError: errors.New("Fail"), saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestStoredIPReadError should fail because reading the stored IP returns an error")
	}

}

// Rule 3 (DNS forces an update): despite its name, this is NOT a no-op case. The
// stored IP matches the current one, so no local update is needed, but the DNS
// cross-check returns a different IP ("any"), so Run still performs the update and
// must succeed. This is the only test exercising the "DNS differs" branch; the
// genuine no-update case is TestStoredIPAndDNSMatchNoUpdate.
func TestStoredIPMatchesButDNSDiffersTriggersUpdate(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.1", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestStoredIPMatchesButDNSDiffersTriggersUpdate should not fail")
	}

}

// Rule 2: the stored IP differs from the current one, so an update is required.
// The DNS cross-check is skipped (update already decided) and all notifications
// and the save succeed, so Run returns no error.
func TestStoredIPDiffersTriggersUpdate(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.2", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestStoredIPDiffersTriggersUpdate should not fail")
	}

}

// Rule 2: there is no stored IP yet (found == false), so an update is required
// and the full notify + save path runs successfully.
func TestNoStoredIPTriggersUpdate(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.2", storeFound: false, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestNoStoredIPTriggersUpdate should not fail")
	}

}

// Rule 3: the stored IP matches, so Run reaches the DNS cross-check, but the
// resolver fails. Run must propagate that resolver error.
func TestStoredIPMatchesDNSResolveError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: errors.New("Fail")}

	store := ipStoreMock{storedIPValue: "1.1.1.1", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestStoredIPMatchesDNSResolveError should fail")
	}

}

// Rule 3 (no-update branch): the stored IP matches AND the DNS record returns the
// same IP, so nothing changed and Run finishes without notifying or saving.
func TestStoredIPAndDNSMatchNoUpdate(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "1.1.1.1", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.1", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err != nil {
		t.Errorf("TestStoredIPAndDNSMatchNoUpdate should not fail")
	}

}

// Rule 4: an update is required (stored IP differs) but the first notification
// (NotifyQueue) fails. Run must return its error before persisting.
func TestUpdateNotifyChangeError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.2", storeFound: true, storeError: nil, saveError: nil}

	notifier := notifierMock{err: errors.New("Fail")}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestUpdateNotifyChangeError should fail, because notify IP change fails")
	}

}

// Rule 4: an update is required and the first notification (NotifyQueue) succeeds
// but the second one (UpdateQueue) fails. complexNotifierMock fails only the
// "update" queue, so this isolates the second notification error.
func TestUpdateNotifyUpdateQueueError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.2", storeFound: true, storeError: nil, saveError: nil}

	notifier := complexNotifierMock{err: errors.New("Fail"), failQueue: "update"}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestUpdateNotifyUpdateQueueError should fail, because notify IP update fails")
	}

}

// Rule 4: an update is required and both notifications succeed, but persisting
// the new IP (SaveIP) fails. Run must return the save error.
func TestUpdateSaveIPError(t *testing.T) {

	ipinfoData := domain.IPInfo{IP: "1.1.1.1", OrgName: "Test"}
	ipinfo :=
		ipInfoMock{ipInfoData: ipinfoData, err: nil}

	resolver := dnsResolverMock{result: "any", err: nil}

	store := ipStoreMock{storedIPValue: "1.1.1.2", storeFound: true, storeError: nil, saveError: errors.New("Fail")}

	notifier := notifierMock{err: nil}

	settings := Settings{ISPName: "Test", DomainName: "test.windmaker.net", NotifyQueue: "notify", UpdateQueue: "update"}

	monitor := NewMonitor(ipinfo, resolver, store, notifier, settings)

	ctx := context.Background()

	err := monitor.Run(ctx)

	if err == nil {
		t.Errorf("TestUpdateSaveIPError should fail, because store fails when IP is saved")
	}

}
