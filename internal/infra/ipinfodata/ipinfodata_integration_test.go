//go:build integration_tests || unit_tests || ipinfodata_tests || ipinfodata_integration_tests

package ipinfodata

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestIPInfoRequester(t *testing.T) {

	client := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	requester := IPInfoRequester{HttpClient: &client}
	ctx := context.Background()
	_, ipInfoErr := requester.GetIPInfo(ctx)
	if ipInfoErr != nil {
		t.Errorf("IPInfoRequester with real client should not fail, error was \"%s\".", ipInfoErr.Error())
	}
}
