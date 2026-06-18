package storage

import (
	"context"

	memorydatabase "github.com/a-castellano/go-services/memorydatabase"
)

// CheckDatabase Checks stored value in memorydatabase and compares its value with ip arg
// in order to decide if value update is required
//
// It compares the provided IP address with the previously stored IP address
// to determine if an update is needed. If no IP is stored (first run),
// it will require an update.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - ip: Current IP address to compare
//   - databaseClient: Interface for database operations
//
// Returns:
//   - bool: True if update is required, false otherwise
//   - error: Error if database operation fails
func CheckDatabase(ctx context.Context, ip string, databaseClient memorydatabase.MemoryDatabase) (bool, error) {

	var requireUpdate bool = false

	// Read the currently stored IP address from Redis
	currentRegisteredValue, found, readErr := databaseClient.ReadString(ctx, "storedIP")

	if readErr != nil {
		return requireUpdate, readErr
	}

	// If no IP is stored (first run), require update
	if found == false {
		requireUpdate = true
	} else {
		// If stored IP differs from current IP, require update
		if currentRegisteredValue != ip {
			requireUpdate = true
		}
	}

	return requireUpdate, nil
}

// UpdateIP updates "storedIP" value in memorydatabase
// It stores the new IP address in Redis for future comparisons
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - ip: IP address to store
//   - databaseClient: Interface for database operations
//
// Returns:
//   - error: Error if database operation fails
func UpdateIP(ctx context.Context, ip string, databaseClient memorydatabase.MemoryDatabase) error {
	// Store IP with no TTL (persistent storage)
	writeError := databaseClient.WriteString(ctx, "storedIP", ip, 0)
	return writeError
}
