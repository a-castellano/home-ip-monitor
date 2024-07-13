package ipinfo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// IPinfo stores ingo retrieved from https://ipinfo.io/ calls
type IPinfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
	Readme   string `json:"readme"`
	OrgName  string
}

// getOrgName retrieves OrgName from Org field
func (ipInfo *IPinfo) getOrgName() {
	ipInfo.OrgName = strings.Split(ipInfo.Org, " ")[1]
	log.Printf("Retrieved OrgName is \"%s\"", ipInfo.OrgName)
}

// Requester interface must implement GetIPInfo method
type Requester interface {
	GetIPInfoResponse() (*http.Response, error)
}

// RetireveIPInfoFromResponse Proceses GetIPInfoResponse response and return IPinfo data
func RetireveIPInfoFromResponse(requester Requester) (IPinfo, error) {

	var retrievedInfo IPinfo

	response, responseError := requester.GetIPInfoResponse()

	if responseError != nil {
		return retrievedInfo, responseError
	}

	defer response.Body.Close()
	bs, _ := ioutil.ReadAll(response.Body)

	unmarshalErr := json.Unmarshal(bs, &retrievedInfo)
	if unmarshalErr != nil {
		return retrievedInfo, unmarshalErr
	}

	if retrievedInfo.IP == "" {
		return retrievedInfo, errors.New("no IPInfo was found during request phase")
	}

	log.Printf("Retrieved IP is \"%s\"", retrievedInfo.IP)
	retrievedInfo.getOrgName()

	return retrievedInfo, nil

}

// Realrequester is the actual requester
type Realrequester struct {
	Client http.Client
}

// GetIPInfoResponse retrievesa ctual ipinfo response
func (requester Realrequester) GetIPInfoResponse() (*http.Response, error) {
	request, _ := http.NewRequest("GET", "https://ipinfo.io/", nil)
	response, responseError := requester.Client.Do(request)

	return response, responseError
}
