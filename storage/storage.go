package storage

import (
	"context"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
)

// CheckDatabase Checks stored value in memorydatabase and compares its value with ip arg
// in order todecide if value update is required
func CheckDatabase(ctx context.Context, ip string, databaseClient memorydatabase.MemoryDatabase) (bool, error) {

	var requireUpdate bool = false

	currentRegisteredValue, found, readErr := databaseClient.ReadString(ctx, "storedIP")

	if readErr != nil {
		return requireUpdate, readErr
	}

	if found == false {
		requireUpdate = true
	} else {
		if currentRegisteredValue != ip {
			requireUpdate = true
		}
	}

	return requireUpdate, nil
}

// UpdateIP updates "storedIP" value in memorydatabase
func UpdateIP(ctx context.Context, ip string, databaseClient memorydatabase.MemoryDatabase) error {
	// ttl is 0
	writeError := databaseClient.WriteString(ctx, "storedIP", ip, 0)
	return writeError
}
