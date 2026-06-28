package dns

type Provider interface {
	GetCurrentIP(domain, name string) (string, error)
	UpsertARecord(domain, name, ip string, ttl int) error
	DeleteRecord(domain, name string) error
}
