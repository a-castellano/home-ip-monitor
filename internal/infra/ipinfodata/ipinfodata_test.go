//go:build integration_tests || unit_tests || ipinfodata_tests || ipinfodata_unit_tests

package ipinfodata

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	mock "github.com/a-castellano/home-ip-monitor/internal/infra/ipinfodata/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetIPInfoRequestCreationError(t *testing.T) {
	// Force NewRequestWithContext to fail: a URL containing a control
	// character (0x7f) makes url.Parse return an error.
	old := ipInfoURL
	ipInfoURL = "https://ipinfo.io/\x7f"
	defer func() { ipInfoURL = old }()

	// httpClient is never used: the request fails before reaching Do().
	requester := IPInfoRequester{httpClient: &http.Client{}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when request creation fails")
	}
}

func TestGetIPInfoTransportError(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	transport.EXPECT().
		RoundTrip(gomock.Any()).
		Return(nil, errors.New("boom"))

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when transport fails")
	}
}

func TestGetIPInfoInvalidReturnCodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewBufferString("any content")),
	}, nil)

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when transport fails")
	}
}

// errorReader is an io.ReadCloser whose Read always fails.
type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errorReader) Close() error             { return nil }

func TestGetIPInfoBodyReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	// 200 OK, but the body errors on Read so io.ReadAll fails.
	transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       errorReader{},
	}, nil)

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when the response body cannot be read")
	}
}

func TestGetIPInfoBodyBrokenJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	// 200 OK, but the body JSON response is broken.
	transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`"broken": "json"}`)),
	}, nil)

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when the response body is an invalid JSON")
	}
}

func TestGetIPInfoBodyValidJSONEmptyIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	// 200 OK, but the body JSON response is broken.
	transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"nonsense": "json"}`)),
	}, nil)

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	_, err := requester.GetIPInfo(context.Background())
	if err == nil {
		t.Fatal("GetIPInfo should fail when the response body is an invalid JSON")
	}
}

func TestGetIPInfoTelefonica(t *testing.T) {
	ctrl := gomock.NewController(t)
	transport := mock.NewMockRoundTripper(ctrl)

	// 200 OK, but the body JSON response is broken.
	transport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"ip": "79.12.12.12","hostname": "79-12-12-12.digimobil.es","city": "Madrid","region": "Madrid","country": "ES","loc": "40.4165,-3.7026","org": "AS57269 DIGI SPAIN TELECOM S.L.","postal": "28087","timezone": "Europe/Madrid","readme": "https://ipinfo.io/missingauth"}`)),
	}, nil)

	requester := IPInfoRequester{httpClient: &http.Client{Transport: transport}}

	ipinfo, err := requester.GetIPInfo(context.Background())
	if err != nil {
		t.Fatal("GetIPInfo shouldn't fail when valid JSON is returned")
	} else {
		expectedIP := "79.12.12.12"
		if ipinfo.IP != expectedIP {
			t.Fatalf("ipinfo.IP should be '%s' but got '%s'", expectedIP, ipinfo.IP)
		}
	}
}
