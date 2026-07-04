package storage

import (
	"context"

	logger "github.com/a-castellano/go-services/infra/logger"
	memorydatabase "github.com/a-castellano/go-services/services/memorydatabase"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/a-castellano/home-ip-monitor/internal/infra/storage"

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

	ctx, span := otel.Tracer(tracerName).Start(ctx, "StoredIP",
		trace.WithAttributes(
			attribute.String("operation", "StoredIP"),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "StoredIP")
	log.DebugContext(ctx, "Retrieving stored IP from store")

	value, found, err := store.Database.ReadString(ctx, "storedIP")

	if err != nil {
		errorString := "failed to read stored ip from database"

		// Status only, no RecordError: the error event is recorded by the
		// failing child span (memorydatabase/redis); recording it here again
		// would duplicate the same event up the chain.
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "error", err.Error())
	}

	return value, found, err
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

	ctx, span := otel.Tracer(tracerName).Start(ctx, "SaveIP",
		trace.WithAttributes(
			attribute.String("operation", "SaveIP"),
		),
	)
	defer span.End()

	log := logger.FromContext(ctx).With("operation", "SaveIP")
	log.DebugContext(ctx, "Storing required IP into store", "ip", ip)

	writeError := store.Database.WriteString(ctx, "storedIP", ip, 0)
	if writeError != nil {
		errorString := "failed to store IP into database"

		// Status only, no RecordError: the error event is recorded by the
		// failing child span (memorydatabase/redis); recording it here again
		// would duplicate the same event up the chain.
		span.SetStatus(codes.Error, errorString)

		log.ErrorContext(ctx, errorString, "error", writeError.Error())

	}

	return writeError
}
