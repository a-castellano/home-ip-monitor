package storage

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
)

type IPStore struct {
	database memorydatabase.MemoryDatabase
}

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
func (ipstore *IPStore) StoredIP(ctx context.Context) (string, bool, error) {

	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Retrieving stored IP from ipstore", "operation", "StoredIP")

	return ipstore.database.ReadString(ctx, "storedIP")
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
func (ipstore *IPStore) SaveIP(ctx context.Context, ip string) error {
	// Store IP with no TTL (persistent storage)
	log := logger.FromContext(ctx)
	log.DebugContext(ctx, "Storing required IP into ipstore", "ip", ip, "operation", "SaveIP")

	writeError := ipstore.database.WriteString(ctx, "storedIP", ip, 0)
	return writeError
}
