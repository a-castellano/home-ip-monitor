//go:build integration_tests || unit_tests || ipinfo_tests || ipinfo_unit_tests

package ipinfo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type RoundTripperMock struct {
	Response *http.Response
	RespErr  error
}

func (rtm *RoundTripperMock) RoundTrip(*http.Request) (*http.Response, error) {
	return rtm.Response, rtm.RespErr
}

type MockIPinfo struct {
	provider string
}

func (m MockIPinfo) GetIPInfoResponse() (*http.Response, error) {

	var client http.Client

	switch m.provider {

	case "Digi":
		client = http.Client{Transport: &RoundTripperMock{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{"ip": "79.12.12.12","hostname": "79-12-12-12.digimobil.es","city": "Madrid","region": "Madrid","country": "ES","loc": "40.4165,-3.7026","org": "AS57269 DIGI SPAIN TELECOM S.L.","postal": "28087","timezone": "Europe/Madrid","readme": "https://ipinfo.io/missingauth"}`))}}}

	case "Telefonica":
		client = http.Client{Transport: &RoundTripperMock{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{"ip": "95.12.12.12","hostname": "12.red-95-12-12.staticip.rima-tde.net","city": "Valencia","region": "Valencia","country": "ES","loc": "39.4739,-0.3797","org": "AS3352 TELEFONICA DE ESPANA S.A.U.","postal": "46001","timezone": "Europe/Madrid","readme": "https://ipinfo.io/missingauth"}`))}}}

	case "invalid":
		client = http.Client{Transport: &RoundTripperMock{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`{"nonsense": "json"}`))}}}

	default:
		client = http.Client{Transport: &RoundTripperMock{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString(`"nonsense": "json"}`))}}}
	}

	request, _ := http.NewRequest("GET", "https://ipinfo.io/", nil)
	response, responseError := client.Do(request)

	return response, responseError
}

func TestErroredRequester(t *testing.T) {

	erroredRequester := MockIPinfo{}
	_, ipInfoErr := RetrieveIPInfoFromResponse(erroredRequester)

	if ipInfoErr == nil {
		t.Errorf("TestErroredRequester should fail")
	} else {
		if ipInfoErr.Error() != "invalid character ':' after top-level value" {
			t.Errorf("TestErroredRequester error should be \"invalid character ':' after top-level value\", not \"%s\".", ipInfoErr.Error())
		}
	}

}

func TestInvalidRequester(t *testing.T) {

	invalidRequester := MockIPinfo{provider: "invalid"}
	_, ipInfoErr := RetrieveIPInfoFromResponse(invalidRequester)

	if ipInfoErr == nil {
		t.Errorf("TestInvalidRequester should fail.")
	} else {
		if ipInfoErr.Error() != "no IPInfo was found during request phase" {
			t.Errorf("TestInvalidRequester error should be \"no IPInfo was found during request phase\", not \"%s\".", ipInfoErr.Error())
		}
	}
}

func TestDigiRequester(t *testing.T) {

	digiRequester := MockIPinfo{provider: "Digi"}
	ipInfo, ipInfoErr := RetrieveIPInfoFromResponse(digiRequester)

	if ipInfoErr != nil {
		t.Errorf("TestDigiRequester should not fail, error was \"%s\".", ipInfoErr.Error())
	} else {
		if ipInfo.IP != "79.12.12.12" {
			t.Errorf("ipInfo.IP should be \"79.12.12.12\" not \"%s\".", ipInfo.IP)
		}
		if ipInfo.OrgName != "DIGI" {
			t.Errorf("ipInfo.OrgName should be \"DIGI\" not \"%s\".", ipInfo.OrgName)
		}
	}
}

func TestTelefonicaRequester(t *testing.T) {

	telefonicaRequester := MockIPinfo{provider: "Telefonica"}
	ipInfo, ipInfoErr := RetrieveIPInfoFromResponse(telefonicaRequester)

	if ipInfoErr != nil {
		t.Errorf("TesttelefonicaRequester should not fail, error was \"%s\".", ipInfoErr.Error())
	} else {
		if ipInfo.IP != "95.12.12.12" {
			t.Errorf("ipInfo.IP should be \"95.12.12.12\" not \"%s\".", ipInfo.IP)
		}
		if ipInfo.OrgName != "TELEFONICA" {
			t.Errorf("ipInfo.OrgName should be \"TELEFONICA\" not \"%s\".", ipInfo.OrgName)
		}
	}
}
