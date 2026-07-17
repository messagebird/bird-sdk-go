package bird

import (
	"bytes"
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// DomainsService manages sending domains: register, read, update, delete, list,
// and verify. Reach it via Client.Domains. Register a domain, publish the DNS
// records it returns, then call Verify until it is usable as a sender.
type DomainsService struct{ client *Client }

// DomainSettings toggles per-domain sending behavior. Each field is a pointer:
// nil leaves it unchanged (on update) or defaults server-side (on create);
// point it at a value to set it. Enabling tracking requires a tracking domain,
// else the API returns 409.
type DomainSettings struct {
	ClickTracking *bool
	OpenTracking  *bool
}

// DomainTrackingConfig configures the branded open/click tracking domain.
// Provide only the name part; Bird appends the sending domain (e.g. "links" on
// "mail.acme.com" becomes "links.mail.acme.com").
type DomainTrackingConfig struct {
	Name string
}

// DomainReturnPathConfig configures the return-path (bounce) domain. Provide
// only the name part; Bird appends the sending domain.
type DomainReturnPathConfig struct {
	Name string
}

// DomainDKIMConfig configures DKIM signing. Mode is "" (default), "txt", or
// "delegated" ("delegated" is a preview value the server currently rejects).
type DomainDKIMConfig struct {
	Mode string
}

// DomainInboundConfig enables or disables receiving on the domain.
type DomainInboundConfig struct {
	Enabled bool
}

// DomainCreateParams registers a sending domain. Domain is required; the config
// blocks are optional and default server-side when omitted.
type DomainCreateParams struct {
	Domain     string // required
	ReturnPath *DomainReturnPathConfig
	Tracking   *DomainTrackingConfig
	Dkim       *DomainDKIMConfig
	Settings   *DomainSettings
}

func (p DomainCreateParams) toWire() oapi.DomainCreate {
	body := oapi.DomainCreate{Domain: p.Domain}
	if p.Settings != nil {
		body.Settings = &oapi.DomainSettings{ClickTracking: p.Settings.ClickTracking, OpenTracking: p.Settings.OpenTracking}
	}
	if p.Tracking != nil {
		body.Tracking = &oapi.DomainTrackingConfig{Name: p.Tracking.Name}
	}
	if p.ReturnPath != nil {
		body.ReturnPath = &oapi.DomainReturnPathConfig{Name: p.ReturnPath.Name}
	}
	if p.Dkim != nil && p.Dkim.Mode != "" {
		mode := oapi.DomainDKIMConfigMode(p.Dkim.Mode)
		body.Dkim = &oapi.DomainDKIMConfig{Mode: &mode}
	}
	return body
}

// DomainUpdateParams is a partial update. Every field is optional: a nil config
// leaves that part unchanged. Set ClearTracking to remove the tracking domain
// (sends tracking: null) — both tracking toggles must be off first, else the
// API returns 409. ClearTracking takes precedence over Tracking.
type DomainUpdateParams struct {
	Settings      *DomainSettings
	ReturnPath    *DomainReturnPathConfig
	Tracking      *DomainTrackingConfig
	ClearTracking bool
	Dkim          *DomainDKIMConfig
	Inbound       *DomainInboundConfig
}

func (p DomainUpdateParams) toWire() oapi.DomainUpdate {
	var body oapi.DomainUpdate
	if p.Settings != nil {
		body.Settings = &oapi.DomainSettings{ClickTracking: p.Settings.ClickTracking, OpenTracking: p.Settings.OpenTracking}
	}
	if p.Tracking != nil && !p.ClearTracking {
		body.Tracking = &oapi.DomainTrackingConfig{Name: p.Tracking.Name}
	}
	if p.ReturnPath != nil {
		body.ReturnPath = &struct {
			Name string `json:"name"`
		}{Name: p.ReturnPath.Name}
	}
	if p.Dkim != nil && p.Dkim.Mode != "" {
		mode := oapi.DomainUpdateDkimMode(p.Dkim.Mode)
		body.Dkim = &struct {
			Mode *oapi.DomainUpdateDkimMode `json:"mode,omitempty"`
		}{Mode: &mode}
	}
	if p.Inbound != nil {
		body.Inbound = &struct {
			Enabled bool `json:"enabled"`
		}{Enabled: p.Inbound.Enabled}
	}
	return body
}

// DomainListParams filters the domain list. Zero-value fields are omitted.
type DomainListParams struct {
	Name  string // optional case-insensitive substring match on the domain name
	Limit int
}

func (p DomainListParams) toWire(startingAfter string) *oapi.ListDomainsParams {
	w := &oapi.ListDomainsParams{}
	if p.Name != "" {
		name := p.Name
		w.Name = &name
	}
	if p.Limit > 0 {
		limit := oapi.PaginationLimit(p.Limit)
		w.Limit = &limit
	}
	if startingAfter != "" {
		cursor := oapi.StartingAfter(startingAfter)
		w.StartingAfter = &cursor
	}
	return w
}

// Create registers a sending domain. It returns in "pending" with the DNS
// records to publish; call Verify once they are in place. Retried safely: a
// single idempotency key is reused across attempts. Provide your own key with
// option.WithIdempotencyKey.
func (s *DomainsService) Create(ctx context.Context, params DomainCreateParams, opts ...option.RequestOption) (*Domain, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.CreateDomainParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.CreateDomain(ctx, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Domain
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get returns a single sending domain by id, with its DNS records and their
// per-record verification state.
func (s *DomainsService) Get(ctx context.Context, id string, opts ...option.RequestOption) (*Domain, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.GetDomain(ctx, id, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Domain
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update edits a sending domain. Only the fields set in params change; settings
// apply immediately, while return-path/tracking/DKIM changes are staged until
// their new DNS records verify. Retried safely with a reused idempotency key.
func (s *DomainsService) Update(ctx context.Context, id string, params DomainUpdateParams, opts ...option.RequestOption) (*Domain, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	wire := params.toWire()
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.UpdateDomainParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		// The generated body omits a nil tracking pointer, so an explicit
		// removal (tracking: null) has to go out as raw JSON.
		if params.ClearTracking {
			raw, err := marshalUpdateWithTrackingNull(wire)
			if err != nil {
				return nil, err
			}
			return s.client.oapi.UpdateDomainWithBody(ctx, id, p, "application/json", bytes.NewReader(raw), s.client.callEditors(cfg)...)
		}
		return s.client.oapi.UpdateDomain(ctx, id, p, wire, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Domain
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// marshalUpdateWithTrackingNull serializes the update body and forces an
// explicit tracking: null, which the typed body (omitempty) can't express.
func marshalUpdateWithTrackingNull(w oapi.DomainUpdate) ([]byte, error) {
	raw, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	fields["tracking"] = json.RawMessage("null")
	return json.Marshal(fields)
}

// Delete removes a sending domain. Mail already accepted still sends; no new
// mail can be sent from it. Retried safely with a reused idempotency key.
func (s *DomainsService) Delete(ctx context.Context, id string, opts ...option.RequestOption) error {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return err
	}
	_, err = cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.DeleteDomainParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.DeleteDomain(ctx, id, p, s.client.callEditors(cfg)...)
	})
	return err
}

// Verify triggers a fresh DNS check and returns the refreshed domain with
// per-record results. Safe to repeat while waiting for DNS to propagate.
// Retried safely with a reused idempotency key.
func (s *DomainsService) Verify(ctx context.Context, id string, opts ...option.RequestOption) (*Domain, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, true, func(ctx context.Context, idempotencyKey string) (*http.Response, error) {
		p := &oapi.VerifyDomainParams{}
		if idempotencyKey != "" {
			p.IdempotencyKey = &idempotencyKey
		}
		return s.client.oapi.VerifyDomain(ctx, id, p, s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out Domain
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPage fetches one page of sending domains. Pass the previous page's
// NextCursor as startingAfter to advance; "" starts from the most recent.
func (s *DomainsService) ListPage(ctx context.Context, params DomainListParams, startingAfter string, opts ...option.RequestOption) (*DomainList, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	body, err := cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return s.client.oapi.ListDomains(ctx, params.toWire(startingAfter), s.client.callEditors(cfg)...)
	})
	if err != nil {
		return nil, err
	}
	var out DomainList
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List walks every sending domain matching params, fetching pages lazily. Range
// over it; the second value is non-nil only on the iteration where a fetch
// failed.
func (s *DomainsService) List(ctx context.Context, params DomainListParams, opts ...option.RequestOption) iter.Seq2[*Domain, error] {
	return paginate(func(cursor string) ([]Domain, *string, error) {
		page, err := s.ListPage(ctx, params, cursor, opts...)
		if err != nil {
			return nil, nil, err
		}
		return page.Data, page.NextCursor, nil
	})
}
