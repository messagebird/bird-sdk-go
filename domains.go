package bird

// DomainsService manages sending domains: register, read, update, delete, list,
// and verify. Reach it via Client.Domains. Register a domain, publish the DNS
// records it returns, then call Verify until it is usable as a sender.
type DomainsService struct{ resource }
