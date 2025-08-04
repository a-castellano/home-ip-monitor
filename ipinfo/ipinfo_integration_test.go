//go:build integration_tests || ipinfo_tests || ipinfo_unit_tests

package ipinfo

import (
	"net/http"
	"testing"
	"time"
)

func TestIPInfoRequester(t *testing.T) {

	client := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	requester := Realrequester{Client: client}
	ipInfo, ipInfoErr := RetrieveIPInfoFromResponse(requester)

	if ipInfoErr != nil {
		t.Errorf("TesttelefonicaRequester should not fail, error was \"%s\".", ipInfoErr.Error())
	} else {
		if ipInfo.IP == "" {
			t.Errorf("ipInfo.IP should not be empty.")
		}
		if ipInfo.OrgName == "" {
			t.Errorf("ipInfo.OrgName should not be empty.")
		}
	}
}
