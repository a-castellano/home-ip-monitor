//go:build integration_tests || unit_tests || storage_tests || storage_unit_tests

package storage

import (
	"context"
	"errors"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
	redismock "github.com/go-redis/redismock/v9"
	goredis "github.com/redis/go-redis/v9"
	"testing"
	"time"
)

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

func TestErrorRedis(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetErr(errors.New("FAIL"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	ipstore := Store{database: memoryDatabase}

	_, _, storedIPErr := ipstore.StoredIP(ctx)
	if storedIPErr == nil {
		t.Errorf("TestErrorRedis should return error when redis read has failed.")
	}

}

func TestIPNotSetYetRedis(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").RedisNil()

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)
	ipstore := Store{database: memoryDatabase}

	_, found, storedIPErr := ipstore.StoredIP(ctx)
	if storedIPErr != nil {
		t.Errorf("TestIPNotSetYetRedis shoudld not fail.")
	}
	if found == true {
		t.Errorf("TestIPNotSetYetRedis shouldn't find ay value")
	}

}

func TestStoredIP(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	expectedIP := "12.12.12.12"
	mock.ExpectGet("storedIP").SetVal(expectedIP)

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)
	ipstore := Store{database: memoryDatabase}

	storedIP, found, storedIPErr := ipstore.StoredIP(ctx)
	if storedIPErr != nil {
		t.Errorf("TestStoredSameIP should not fail.")
	}
	if found == false {
		t.Errorf("TestStoredSameIP should find an stored IP.")
	}
	if storedIP != expectedIP {
		t.Fatalf("Stored ip should be '%s' instead of the actual stored '%s'", expectedIP, storedIP)
	}

}

func TestUpdateIPWithNoError(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectSet("storedIP", "12.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	ipstore := Store{database: memoryDatabase}
	errorOnUpdate := ipstore.SaveIP(ctx, "12.12.12.12")
	if errorOnUpdate != nil {
		t.Errorf("TestUpdateIPWithNoError should not fail.")
	}

}

func TestUpdateIPWithError(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectSet("storedIP", "12.12.12.12", 0).SetErr(errors.New("FAIL"))

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	ipstore := Store{database: memoryDatabase}
	errorOnUpdate := ipstore.SaveIP(ctx, "12.12.12.12")
	if errorOnUpdate == nil {
		t.Errorf("TestUpdateIPWithError should fail.")
	}

}
