//go:build integration_tests || unit_tests

package monitor

import (
	"bytes"
	"context"
	"errors"
	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
	messagebroker "github.com/a-castellano/go-services/messagebroker"
	config "github.com/a-castellano/home-ip-monitor/config"
	redismock "github.com/go-redis/redismock/v9"
	goredis "github.com/redis/go-redis/v9"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
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

type RedisClientMock struct {
	client *goredis.Client
}

func (mock RedisClientMock) IsClientInitiated() bool {
	return true
}

func (mock RedisClientMock) WriteString(ctx context.Context, key string, value string, ttl int) error {
	return mock.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

func (mock RedisClientMock) ReadString(ctx context.Context, key string) (string, bool, error) {
	var found bool = true
	value, getError := mock.client.Get(ctx, key).Result()

	if getError != nil {
		found = false
		if getError == goredis.Nil {
			return value, found, nil
		} else {
			return value, found, getError
		}
	}
	return value, found, nil
}

type RabbitmqMock struct {
	LaunchError bool
}

func (client RabbitmqMock) SendMessage(queueName string, message []byte) error {
	if client.LaunchError {
		return errors.New("Error")
	}
	return nil
}

func (client RabbitmqMock) ReceiveMessages(ctx context.Context, queueName string, messages chan<- []byte, errorsChan chan<- error) {
	if client.LaunchError {
		errorsChan <- errors.New("Error")
	} else {
		okMessage := []byte("This is ok")
		messages <- okMessage
		errorsChan <- nil
	}
}

type MockResolver struct {
	Response      string
	ResponseError error
}

func (mock MockResolver) GetIP(ctx context.Context, domain string) (string, error) {
	return mock.Response, mock.ResponseError
}

func TestIPAlreadyInDatabase(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Digi"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.12")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}

	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError != nil {
		t.Errorf("TestIPAlreadyInDatabase should not fail.")
	}
}

func TestIPInDatabaseNotSameAsReaded(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Digi"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.11")
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}

	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError != nil {
		t.Errorf("TestIPAlreadyInDatabase should not fail, error was \"%s\".", monitorError.Error())
	}
}

func TestIPDifferentISP(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Telefonica"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.11")
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}

	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError != nil {
		t.Errorf("TestIPAlreadyInDatabase should not fail, error was \"%s\".", monitorError.Error())
	}
}

func TestInvalidISP(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "invalid"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.11")
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}
	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError == nil {
		t.Errorf("TestIPAlreadyInDatabase should fail.")
	}
}

func TestErrorRedisSet(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Digi"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.11")
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetErr(errors.New("FAIL"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}
	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError == nil {
		t.Errorf("TestErrorRedisSet should fail.")
	}
}

func TestErrorRedisGet(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Digi"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetErr(errors.New("FAIL"))
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: false}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}
	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError == nil {
		t.Errorf("TestErrorRedisSet should fail.")
	}
}

func TestNotifyError(t *testing.T) {

	appConfig := config.Config{ISPName: "DIGI", DomainName: "test.windmaker.net"}

	digiRequester := MockIPinfo{provider: "Digi"}

	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("79.12.12.11")
	mock.ExpectSet("storedIP", "79.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	rabbitmock := RabbitmqMock{LaunchError: true}
	messageBroker := messagebroker.MessageBroker{Client: rabbitmock}

	nsLookup := MockResolver{Response: "79.12.12.12", ResponseError: nil}

	monitorError := Monitor(ctx, digiRequester, nsLookup, memoryDatabase, messageBroker, &appConfig)

	if monitorError == nil {
		t.Errorf("TestIPAlreadyInDatabase should fail.")
	}
}
