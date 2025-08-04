//go:build integration_tests || unit_tests

package config

import (
	"os"
	"testing"
)

var currentDomainName string
var currentDomainNameDefined bool

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

	if envDomainName, found := os.LookupEnv("DOMAIN_NAME"); found {
		currentDomainName = envDomainName
		currentDomainNameDefined = true
	} else {
		currentDomainNameDefined = false
	}

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

	if currentDomainNameDefined {
		os.Setenv("DOMAIN_NAME", currentDomainName)
	} else {
		os.Unsetenv("DOMAIN_NAME")
	}

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

func TestConfigWithoutEnvVariables(t *testing.T) {

	setUp()
	defer teardown()

	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithoutEnvVariables should fail.")
	} else {
		if err.Error() != "env variable DOMAIN_NAME must be set" {
			t.Errorf("TestConfigWithoutEnvVariables error should be \"env variable DOMAIN_NAME must be set\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfigWithInvalidRedisPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	os.Setenv("DOMAIN_NAME", "test.windmaker.net")
	os.Setenv("REDIS_PORT", "invalidport")
	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidRedisPort should fail.")
	} else {
		if err.Error() != "strconv.Atoi: parsing \"invalidport\": invalid syntax" {
			t.Errorf("TestConfigWithInvalidRedisPort error should be \"strconv.Atoi: parsing \"invalidport\": invalid syntax\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfigWithInvalidRabbitmqPort(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	os.Setenv("DOMAIN_NAME", "test.windmaker.net")
	os.Setenv("RABBITMQ_PORT", "invalidport")
	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigWithInvalidRabbitmqPort should fail.")
	} else {
		if err.Error() != "strconv.Atoi: parsing \"invalidport\": invalid syntax" {
			t.Errorf("TestConfigWithInvalidRabbitmqPort error should be \"strconv.Atoi: parsing \"invalidport\": invalid syntax\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfigwithoutDNSServer(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DOMAIN_NAME", "test.windmaker.net")
	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigwithoutDNSServer should fail.")
	} else {
		if err.Error() != "env variable DNS_SERVER must be set" {
			t.Errorf("TestConfigwithoutDNSServer error should be \"env variable DNS_SERVER must be set\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfigwithoutDomainName(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigwithoutDomainName should fail.")
	} else {
		if err.Error() != "env variable DOMAIN_NAME must be set" {
			t.Errorf("TestConfigwithoutDomainName error should be \"env variable DOMAIN_NAME must be set\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfigwithoutISPName(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("DOMAIN_NAME", "test.windmaker.net")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	_, err := NewConfig()

	if err == nil {
		t.Errorf("TestConfigwithoutISPName should fail.")
	} else {
		if err.Error() != "env variable ISP_NAME must be set" {
			t.Errorf("TestConfigwithoutISPName error should be \"env variable ISP_NAME must be set\" but it was \"%s\".", err.Error())
		}
	}

}

func TestConfig(t *testing.T) {

	setUp()
	defer teardown()

	os.Setenv("ISP_NAME", "DIGI")
	os.Setenv("DNS_SERVER", "1.1.1.1:53")
	os.Setenv("DOMAIN_NAME", "test.windmaker.net")
	config, err := NewConfig()

	if err != nil {
		t.Errorf("TestConfigWithoutEnvVariables should not fail.")
	} else {
		if config.ISPName != "DIGI" {
			t.Errorf("config.ISPName \"DIGI\" but it was \"%s\".", config.ISPName)
		}
		if config.UpdateQueue != "home-ip-monitor-updates" {
			t.Errorf("config.UpdateQueue \"home-ip-monitor-updates\" but it was \"%s\".", config.UpdateQueue)
		}
		if config.NotifyQueue != "home-ip-monitor-notifications" {
			t.Errorf("config.NotifyQueue \"home-ip-monitor-notifications\" but it was \"%s\".", config.NotifyQueue)
		}

	}

}
