//go:build integration_tests || unit_tests || ipinfodata_tests || ipinfodata_integration_tests

package ipinfodata

import (
	"contex"
	"net/http"
	"testing"
	"time"
)

func TestIPInfoRequester(t *testing.T) {

	client := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	requester := IPInfoRequester{Client: client}
	ctx := context.Background()
	_, ipInfoErr := RetrieveIPInfoFromResponse(requester)
	if ipInfoErr != nil {
		t.Errorf("IPInfoRequester with real client should not fail, error was \"%s\".", ipInfoErr.Error())
	} else {

		if ipInfo.OrgName == "DIGI" {
			t.Errorf("ipInfo.OrgName should not be DIGI.")
		}
	}
}
