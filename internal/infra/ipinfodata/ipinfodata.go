package ipinfodata

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

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
func (ipinfoData ipinfoData) getOrgName(ctx context.Context) domain.IPInfo {

	log := logger.FromContext(ctx)

	orgName := strings.Split(ipinfoData.Org, " ")[1]

	ipinfo := domain.IPInfo{IP: ipinfoData.IP, OrgName: orgName}

	log.InfoContext(ctx, "Retrieve IPInfo data", "data", ipinfo, "operation", "getOrgName")
	return ipinfo
}

type IPInfoRequester struct {
	httpClient *http.Client
}

var (
	ipInfoURL string = "https://ipinfo.io/"
)

func (requester IPInfoRequester) GetIPInfo(ctx context.Context) (domain.IPInfo, error) {
	ipinfo := domain.IPInfo{}
	var retrievedInfo ipinfoData

	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Creating a request to ipinfo", "url", ipInfoURL, "operation", "GetIPInfo")
	req, reqErr := http.NewRequestWithContext(ctx, "GET", ipInfoURL, nil)

	if reqErr != nil {
		log.ErrorContext(ctx, "Error during request to ipinfo creation", "url", ipInfoURL, "error", reqErr.Error(), "operation", "GetIPInfo")
		return ipinfo, reqErr
	}

	log.DebugContext(ctx, "Executing request to ipinfo", "url", ipInfoURL, "operation", "GetIPInfo")

	response, responseErr := requester.httpClient.Do(req)

	if responseErr != nil {
		log.ErrorContext(ctx, "Error performing request to ipinfo", "url", ipInfoURL, "error", responseErr.Error(), "operation", "GetIPInfo")
		return ipinfo, responseErr
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.ErrorContext(ctx, "Error performing request to ipinfo, returned status code is nt 200", "url", ipInfoURL, "StatusCode", response.StatusCode, "operation", "GetIPInfo")

		return ipinfo, errors.New("error performing request to ipinfo, returned status code is not 200")
	}

	log.DebugContext(ctx, "Reading body response", "operation", "GetIPInfo")
	body, bodyErr := io.ReadAll(response.Body)
	if bodyErr != nil {
		log.ErrorContext(ctx, "Error reading body response from ipinfo", "url", ipInfoURL, "error", bodyErr, "operation", "GetIPInfo")
		return ipinfo, bodyErr
	}

	log.DebugContext(ctx, "Parsing JSON body response", "operation", "GetIPInfo")
	unmarshalErr := json.Unmarshal(body, &retrievedInfo)
	if unmarshalErr != nil {
		log.ErrorContext(ctx, "Error reading json response from ipinfo", "url", ipInfoURL, "error", unmarshalErr, "operation", "GetIPInfo")
		return ipinfo, unmarshalErr
	}

	if retrievedInfo.IP == "" {
		log.ErrorContext(ctx, "Error processing json response from ipinfo, no ip has been returned", "operation", "GetIPInfo")
		return ipinfo, errors.New("no IP was returned by ipinfo was found during request")
	}

	log.DebugContext(ctx, "IPInfo request succeded", "retrievedInfo", retrievedInfo, "operation", "GetIPInfo")
	ipinfo = retrievedInfo.getOrgName(ctx)

	return ipinfo, nil
}
