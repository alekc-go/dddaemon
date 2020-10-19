package dddaemon

// Provider provides a common interface for all providers.
type Provider interface {
	UpdateRecord(ip string) error
}
