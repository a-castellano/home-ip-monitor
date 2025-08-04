package ipinfo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// IPinfo stores info retrieved from https://ipinfo.io/ calls
// It contains all the information returned by the ipinfo.io API
type IPinfo struct {
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
	OrgName  string // Extracted ISP name (e.g., "DIGI")
}

// getOrgName retrieves OrgName from Org field
// It parses the "org" field which typically contains "AS12345 ISP_NAME"
// and extracts just the ISP name part
func (ipInfo *IPinfo) getOrgName() {
	ipInfo.OrgName = strings.Split(ipInfo.Org, " ")[1]
	log.Printf("Retrieved OrgName is \"%s\"", ipInfo.OrgName)
}

// Requester interface must implement GetIPInfo method
// This interface allows for easy testing by mocking the HTTP client
type Requester interface {
	GetIPInfoResponse() (*http.Response, error)
}

// RetrieveIPInfoFromResponse Processes GetIPInfoResponse response and return IPinfo data
// It handles the HTTP response from ipinfo.io, parses the JSON, and extracts the ISP name
//
// Parameters:
//   - requester: Interface for making HTTP requests to ipinfo.io
//
// Returns:
//   - IPinfo: Parsed IP information including ISP details
//   - error: Error if request fails or response is invalid
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
type Realrequester struct {
	Client http.Client
}

// GetIPInfoResponse retrieves actual ipinfo response
// It makes a GET request to ipinfo.io and returns the HTTP response
//
// Returns:
//   - *http.Response: HTTP response from ipinfo.io
//   - error: Error if request fails
func (requester Realrequester) GetIPInfoResponse() (*http.Response, error) {
	request, _ := http.NewRequest("GET", "https://ipinfo.io/", nil)
	response, responseError := requester.Client.Do(request)

	return response, responseError
}
