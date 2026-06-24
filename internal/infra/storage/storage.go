package storage

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
)

// Store is the persistence adapter for the monitored IP. It wraps the
// memorydatabase.MemoryDatabase abstraction (not Redis directly) and
// implements domain.IPStore.
type Store struct {
	Database memorydatabase.MemoryDatabase
}

// StoredIP returns the IP currently persisted under the "storedIP" key.
// It only reads the value; deciding whether an update is required is the
// use case's responsibility, not the store's.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//
// Returns:
//   - string: The stored IP address (empty if none was found)
//   - bool: Whether a value was found
//   - error: Error if the read operation fails
func (store *Store) StoredIP(ctx context.Context) (string, bool, error) {

	log := logger.FromContext(ctx).With("operation", "StoredIP")
	log.DebugContext(ctx, "Retrieving stored IP from store")

	return store.Database.ReadString(ctx, "storedIP")
}

// SaveIP persists ip under the "storedIP" key with no TTL (persistent),
// overwriting any previous value.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - ip: IP address to store
//
// Returns:
//   - error: Error if the write operation fails
func (store *Store) SaveIP(ctx context.Context, ip string) error {
	// Store IP with no TTL (persistent storage)
	log := logger.FromContext(ctx).With("operation", "SaveIP")
	log.DebugContext(ctx, "Storing required IP into store", "ip", ip)

	writeError := store.Database.WriteString(ctx, "storedIP", ip, 0)
	return writeError
}
