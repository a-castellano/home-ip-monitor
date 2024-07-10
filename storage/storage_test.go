//go:build integration_tests || unit_tests || ipinfo_tests || ipinfo_unit_tests

package storage

import (
	"context"
	"errors"
	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
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

	_, errorOnCheck := CheckDatabase(ctx, "12.12.12.12", memoryDatabase)
	if errorOnCheck == nil {
		t.Errorf("TestErrorRedis should return error when redis read has failed.")
	}

}

func TestIPNotSetYetRedis(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").RedisNil()

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	requireUpdate, errorOnCheck := CheckDatabase(ctx, "12.12.12.12", memoryDatabase)
	if errorOnCheck != nil {
		t.Errorf("TestIPNotSetYetRedis shoudld not fail.")
	}
	if requireUpdate == false {
		t.Errorf("TestIPNotSetYetRedis should require update as key\"storedIP\" has been set yet")
	}

}

func TestStoredIPDiffers(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("12.12.12.13")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	requireUpdate, errorOnCheck := CheckDatabase(ctx, "12.12.12.12", memoryDatabase)
	if errorOnCheck != nil {
		t.Errorf("TestStoredIPDiffers should not not fail.")
	}
	if requireUpdate == false {
		t.Errorf("TestStoredIPDiffers should require update as key\"storedIP\" is diferent from new one.")
	}

}

func TestStoredSameIP(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectGet("storedIP").SetVal("12.12.12.12")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	requireUpdate, errorOnCheck := CheckDatabase(ctx, "12.12.12.12", memoryDatabase)
	if errorOnCheck != nil {
		t.Errorf("TestStoredSameIP should not fail.")
	}
	if requireUpdate == true {
		t.Errorf("TestStoredSameIP should not require update as key\"storedIP\" has the same value as new IP.")
	}

}

func TestUpdateIPWithNoError(t *testing.T) {
	ctx := context.Background()
	dbMock, mock := redismock.NewClientMock()
	mock.ExpectSet("storedIP", "12.12.12.12", 0).SetVal("OK")

	redisClientMock := RedisClientMock{client: dbMock}
	memoryDatabase := memorydatabase.NewMemoryDatabase(redisClientMock)

	errorOnUpdate := UpdateIP(ctx, "12.12.12.12", memoryDatabase)
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

	errorOnUpdate := UpdateIP(ctx, "12.12.12.12", memoryDatabase)
	if errorOnUpdate == nil {
		t.Errorf("TestUpdateIPWithError should fail.")
	}

}
