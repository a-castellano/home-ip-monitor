package domain

import "context"

type IPInfoProvider interface {
	GetIPInfo(ctx context.Context) (IPInfo, error)
}
type DNSResolver interface {
	Resolve(ctx context.Context, domain string) (string, error)
}
type IPStore interface {
	StoredIP(ctx context.Context) (ip string, found bool, err error)
	SaveIP(ctx context.Context, ip string) error
}
type Notifier interface {
	Notify(ctx context.Context, queue string, message []byte) error
}
