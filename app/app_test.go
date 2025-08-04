//go:build integration_tests || unit_tests

package app

import (
	"bytes"
	"context"
	config "github.com/a-castellano/home-ip-monitor/config"
	nslookup "github.com/a-castellano/home-ip-monitor/nslookup"
	"io/ioutil"
	"net/http"
	"os"
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

var currentISPName string
var currentISPNameDefined bool

var currentDNSServer string
var currentDNSServerDefined bool

var currentUpdateQueue string
var currentUpdateQueueDefined bool

var currentNotifyQueue string
var currentNotifyQueueDefined bool

var currentRedisHost string
var currentRedisHostDefined bool

var currentRedisPort string
var currentRedisPortDefined bool

var currentRedisDatabase string
var currentRedisDatabaseDefined bool

var currentRedisPassword string
var currentRedisPasswordDefined bool

var currentRabbitmqHost string
var currentRabbitmqHostDefined bool

var currentRabbitmqPort string
var currentRabbitmqPortDefined bool

var currentRabbitmqUser string
var currentRabbitmqUserDefined bool

var currentRabbitmqPassword string
var currentRabbitmqPasswordDefined bool

func setUp() {

	if envISPName, found := os.LookupEnv("ISP_NAME"); found {
		currentISPName = envISPName
		currentISPNameDefined = true
	} else {
		currentISPNameDefined = false
	}

	if envDNSServer, found := os.LookupEnv("DNS_SERVER"); found {
		currentDNSServer = envDNSServer
		currentDNSServerDefined = true
	} else {
		currentDNSServerDefined = false
	}

	if envUpdateQueue, found := os.LookupEnv("UPDATE_QUEUE_NAME"); found {
		currentUpdateQueue = envUpdateQueue
		currentUpdateQueueDefined = true
	} else {
		currentUpdateQueueDefined = false
	}

	if envUpdateQueue, found := os.LookupEnv("UPDATE_QUEUE_NAME"); found {
		currentUpdateQueue = envUpdateQueue
		currentUpdateQueueDefined = true
	} else {
		currentUpdateQueueDefined = false
	}

	if envNotifyQueue, found := os.LookupEnv("NOTIFY_QUEUE_NAME"); found {
		currentNotifyQueue = envNotifyQueue
		currentNotifyQueueDefined = true
	} else {
		currentNotifyQueueDefined = false
	}

	if envRedisPort, found := os.LookupEnv("REDIS_PORT"); found {
		currentRedisPort = envRedisPort
		currentRedisPortDefined = true
	} else {
		currentRedisPortDefined = false
	}

	if envRedisDatabase, found := os.LookupEnv("REDIS_DATABASE"); found {
		currentRedisDatabase = envRedisDatabase
		currentRedisDatabaseDefined = true
	} else {
		currentRedisDatabaseDefined = false
	}

	if envRedisPassword, found := os.LookupEnv("REDIS_PASSWORD"); found {
		currentRedisPassword = envRedisPassword
		currentRedisPasswordDefined = true
	} else {
		currentRedisPasswordDefined = false
	}

	if envRabbitmqHost, found := os.LookupEnv("RABBITMQ_HOST"); found {
		currentRabbitmqHost = envRabbitmqHost
		currentRabbitmqHostDefined = true
	} else {
		currentRabbitmqHostDefined = false
	}

	if envRabbitmqPort, found := os.LookupEnv("RABBITMQ_PORT"); found {
		currentRabbitmqPort = envRabbitmqPort
		currentRabbitmqPortDefined = true
	} else {
		currentRabbitmqPortDefined = false
	}

	if envRabbitmqUser, found := os.LookupEnv("RABBITMQ_USER"); found {
		currentRabbitmqUser = envRabbitmqUser
		currentRabbitmqUserDefined = true
	} else {
		currentRabbitmqUserDefined = false
	}

	if envRabbitmqPassword, found := os.LookupEnv("RABBITMQ_PASSWORD"); found {
		currentRabbitmqPassword = envRabbitmqPassword
		currentRabbitmqPasswordDefined = true
	} else {
		currentRabbitmqPasswordDefined = false
	}

	os.Unsetenv("ISP_NAME")
	os.Unsetenv("DNS_SERVER")
	os.Unsetenv("UPDATE_QUEUE_NAME")
	os.Unsetenv("NOTIFY_QUEUE_NAME")

	os.Unsetenv("REDIS_HOST")
	os.Unsetenv("REDIS_PORT")
	os.Unsetenv("REDIS_DATABASE")
	os.Unsetenv("REDIS_PASSWORD")

	os.Unsetenv("RABBITMQ_HOST")
	os.Unsetenv("RABBITMQ_PORT")
	os.Unsetenv("RABBITMQ_DATABASE")
	os.Unsetenv("RABBITMQ_PASSWORD")

}

func teardown() {

	if currentISPNameDefined {
		os.Setenv("ISP_NAME", currentISPName)
	} else {
		os.Unsetenv("ISP_NAME")
	}

	if currentDNSServerDefined {
		os.Setenv("DNS_SERVER", currentDNSServer)
	} else {
		os.Unsetenv("DNS_SERVER")
	}

	if currentUpdateQueueDefined {
		os.Setenv("UPDATE_QUEUE_NAME", currentUpdateQueue)
	} else {
		os.Unsetenv("UPDATE_QUEUE_NAME")
	}

	if currentNotifyQueueDefined {
		os.Setenv("NOTIFY_QUEUE_NAME", currentNotifyQueue)
	} else {
		os.Unsetenv("NOTIFY_QUEUE_NAME")
	}

	if currentRedisHostDefined {
		os.Setenv("REDIS_HOST", currentRedisHost)
	} else {
		os.Unsetenv("REDIS_HOST")
	}

	if currentRedisPortDefined {
		os.Setenv("REDIS_PORT", currentRedisPort)
	} else {
		os.Unsetenv("REDIS_PORT")
	}

	if currentRedisDatabaseDefined {
		os.Setenv("REDIS_DATABASE", currentRedisDatabase)
	} else {
		os.Unsetenv("REDIS_DATABASE")
	}

	if currentRedisPasswordDefined {
		os.Setenv("REDIS_PASSWORD", currentRedisPassword)
	} else {
		os.Unsetenv("REDIS_PASSWORD")
	}

	if currentRabbitmqHostDefined {
		os.Setenv("RABBITMQ_HOST", currentRabbitmqHost)
	} else {
		os.Unsetenv("RABBITMQ_HOST")
	}

	if currentRabbitmqPortDefined {
		os.Setenv("RABBITMQ_PORT", currentRabbitmqPort)
	} else {
		os.Unsetenv("RABBITMQ_PORT")
	}

	if currentRabbitmqUserDefined {
		os.Setenv("RABBITMQ_USER", currentRabbitmqUser)
	} else {
		os.Unsetenv("RABBITMQ_USER")
	}

	if currentRabbitmqPasswordDefined {
		os.Setenv("RABBITMQ_PASSWORD", currentRabbitmqPassword)
	} else {
		os.Unsetenv("RABBITMQ_PASSWORD")
	}

}

func TestIPOK(t *testing.T) {
	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	os.Setenv("REDIS_HOST", "valkey")
	os.Setenv("RABBITMQ_HOST", "rabbitmq")
	os.Setenv("DOMAIN_NAME", "test.windmaker.net")

	appConfig, configErr := config.NewConfig()

	if configErr != nil {
		t.Errorf("TestIPOK should not fail, but it did with error: %s", configErr.Error())
	} else {

		digiRequester := MockIPinfo{provider: "Digi"}

		ctx := context.Background()

		nsLookup := nslookup.DNSLookup{DNSServer: appConfig.DNSServer}

		monitorError := Monitor(ctx, digiRequester, nsLookup, appConfig)

		if monitorError != nil {
			t.Errorf("TestIPOK should not fail.")
		}
	}
}
