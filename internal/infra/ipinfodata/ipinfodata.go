package ipinfodata

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	logger "github.com/a-castellano/go-services/infra/logger"
	domain "github.com/a-castellano/home-ip-monitor/internal/domain"
)

// IPinfo stores info retrieved from https://ipinfo.io/ calls
// It contains all the information returned by the ipinfo.io API
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
func (ipinfoData *ipinfoData) getOrgName(ctx context.Context) domain.IPInfo {

	log := logger.FromContext(ctx)

	orgName := strings.Split(ipinfoData.Org, " ")[1]

	ipinfo := domain.IPInfo{IP: ipinfoData.IP, OrgName: orgName}

	log.InfoContext(ctx, "Retrieve IPInfo data", "data", ipinfo, "operation", "getOrgName")
	return ipinfo
}

	GetIPInfo(ctx context.Context) (domain.IPInfo, error)
}

func RetrieveIPInfoFromResponse(requester Requester) (IPinfo, error) {

	var retrievedInfo IPinfo

	// Make HTTP request to ipinfo.io
	response, responseError := requester.GetIPInfoResponse()

	if responseError != nil {
		return retrievedInfo, responseError
	}

	// Read and parse response body
	defer response.Body.Close()
	bs, _ := ioutil.ReadAll(response.Body)

	// Parse JSON response into IPinfo struct
	unmarshalErr := json.Unmarshal(bs, &retrievedInfo)
	if unmarshalErr != nil {
		return retrievedInfo, unmarshalErr
	}

	// Validate that we received a valid IP address
	if retrievedInfo.IP == "" {
		return retrievedInfo, errors.New("no IPInfo was found during request phase")
	}

	log.Printf("Retrieved IP is \"%s\"", retrievedInfo.IP)
	// Extract ISP name from the organization field
	retrievedInfo.getOrgName()

	return retrievedInfo, nil

}

// Realrequester is the actual requester implementation
// It uses a real HTTP client to make requests to ipinfo.io
type InfoRequester struct {
	Client http.Client
}

func (requester InfoRequester) getIPInfoResponse() (*http.Response, error) {
	request, _ := http.NewRequest("GET", "https://ipinfo.io/", nil)
	response, responseError := requester.Client.Do(request)

	return response, responseError
}

func (requester InfoRequester)   GetIPInfo(ctx context.Context) (domain.IPInfo, error) {
	//to do
}
