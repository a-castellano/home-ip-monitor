package ipinfodata

//go:generate go tool mockgen -destination mocks/http.go -package mock net/http RoundTripper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/a-castellano/home-ip-monitor/internal/infra/ipinfodata"

// ipinfoData is the unexported DTO that maps the raw JSON response from
// ipinfo.io. It holds every wire field, but only IP and Org are used to
// build the domain.IPInfo entity.
type ipinfoData struct {
	IP       string `json:"ip"`       // Public IP address
	Hostname string `json:"hostname"` // Reverse DNS hostname
	City     string `json:"city"`     // City location
	Region   string `json:"region"`   // Region/state
	Country  string `json:"country"`  // Country code
	Loc      string `json:"loc"`      // Latitude/longitude
	Org      string `json:"org"`      // Organization/ISP (e.g., "AS12345 DIGI")
	Postal   string `json:"postal"`   // Postal code
	Timezone string `json:"timezone"` // Timezone
	Readme   string `json:"readme"`   // API documentation URL
}

// getOrgName retrieves OrgName from Org field
// It parses the "org" field which typically contains "AS12345 ISP_NAME"
// and extracts just the ISP name part
func (ipinfoData ipinfoData) getOrgName(ctx context.Context) (domain.IPInfo, error) {

	log := logger.FromContext(ctx).With("operation", "getOrgName")
	var ipinfo domain.IPInfo

	splitedOrgData := strings.Split(ipinfoData.Org, " ")
	if len(splitedOrgData) < 2 {
		log.ErrorContext(ctx, "ipinfo retrieved Org value format cannot be processed", "data", ipinfoData.Org)
		return ipinfo, fmt.Errorf("ipinfo data format cannot be processed : \"%s\"", ipinfoData.Org)
	}
	orgName := splitedOrgData[1]

	ipinfo = domain.IPInfo{IP: ipinfoData.IP, OrgName: orgName}

	log.InfoContext(ctx, "Retrieve IPInfo data", "data", ipinfo)
	return ipinfo, nil
}

// IPInfoRequester is the ipinfo.io HTTP adapter. It implements
// domain.IPInfoProvider by fetching and parsing the public IP information,
// using the injected *http.Client.
type IPInfoRequester struct {
	HttpClient *http.Client
}

// ipInfoURL is the ipinfo.io endpoint queried for public IP information.
// It is a package var (not a const) so tests can override it to exercise
// request-creation errors.
var (
	ipInfoURL string = "https://ipinfo.io/"
)

// GetIPInfo fetches the public IP information from ipinfo.io and maps it to a
// domain.IPInfo. It builds the request, validates the status code, reads and
// parses the JSON body, ensures an IP was returned, and extracts the ISP name.
// It returns an error if any of those steps fails.
func (requester IPInfoRequester) GetIPInfo(ctx context.Context) (domain.IPInfo, error) {
	ipinfo := domain.IPInfo{}
	var retrievedInfo ipinfoData

	ctx, span := otel.Tracer(tracerName).Start(ctx, "GetIPInfo",
		trace.WithAttributes(
			attribute.String("operation", "GetIPInfo"),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "GetIPInfo")
	log.DebugContext(ctx, "Creating a request to ipinfo", "url", ipInfoURL)
	req, reqErr := http.NewRequestWithContext(ctx, "GET", ipInfoURL, nil)

	if reqErr != nil {
		errorString := "error creating request to ipinfo"

		// RecordError stays here: the error is born in this span, no child
		// records it.
		span.RecordError(reqErr)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "url", ipInfoURL, "error", reqErr.Error())
		return ipinfo, reqErr
	}

	log.DebugContext(ctx, "Executing request to ipinfo", "url", ipInfoURL)

	response, responseErr := requester.HttpClient.Do(req)

	if responseErr != nil {
		errorString := "error performing request to ipinfo"

		// Status only: the error event is already recorded by the failing
		// otelhttp child span.
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "url", ipInfoURL, "error", responseErr.Error())
		return ipinfo, responseErr
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {

		errorString := "error performing request to ipinfo, returned status code is not 200"
		statusCodeError := errors.New(errorString)
		// RecordError stays here: the child span marks 4xx/5xx status, but
		// this Go error is born here and no child records it as an event.
		span.RecordError(statusCodeError)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "url", ipInfoURL, "StatusCode", response.StatusCode)

		return ipinfo, statusCodeError
	}

	log.DebugContext(ctx, "Reading body response")
	body, bodyErr := io.ReadAll(response.Body)
	if bodyErr != nil {

		errorString := "error reading body response from ipinfo"

		span.RecordError(bodyErr)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "url", ipInfoURL, "error", bodyErr)
		return ipinfo, bodyErr
	}

	log.DebugContext(ctx, "parsing JSON body response")
	unmarshalErr := json.Unmarshal(body, &retrievedInfo)
	if unmarshalErr != nil {
		errorString := "error reading json response from ipinfo"

		span.RecordError(unmarshalErr)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "url", ipInfoURL, "error", unmarshalErr)
		return ipinfo, unmarshalErr
	}

	if retrievedInfo.IP == "" {

		errorString := "no IP was returned by ipinfo during request"
		noIPError := errors.New(errorString)

		span.RecordError(noIPError)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString)
		return ipinfo, noIPError

	}

	log.DebugContext(ctx, "IPInfo request succeded", "retrievedInfo", retrievedInfo)
	ipinfo, getOrgNameErr := retrievedInfo.getOrgName(ctx)

	if getOrgNameErr != nil {
		errorString := "error processing ipinfo Org name retrieval"

		span.RecordError(getOrgNameErr)
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "error", getOrgNameErr)

		return ipinfo, getOrgNameErr
	}
	span.SetAttributes(attribute.String("ip.retrieved", ipinfo.IP))
	return ipinfo, nil
}
