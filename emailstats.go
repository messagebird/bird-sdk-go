package bird

import (
	"context"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/messagebird/bird-sdk-go/option"
)

// EmailStatsService reads aggregated email statistics. Reach it via
// Client.Email.Stats. Every method is a read; each takes a params struct whose
// fields are all optional (zero values are omitted, and the server applies its
// own defaults for the window, sort, and limit).
type EmailStatsService struct{ client *Client }

// Window and dimension filters shared across the stats endpoints:
//
//   - From, To bound the reporting window. For the day-grained breakdowns and
//     Daily they name calendar days; for Hourly they are instants; for Summary
//     they default to a calendar-day window. A zero value omits the bound and
//     the server fills its default (typically the last 30 days).
//   - Timezone is an IANA identifier (e.g. "America/New_York") the window and
//     bucket boundaries are computed in; empty means UTC.
//   - Category, SendingDomain, Tag, SendingIp, RecipientDomain, and Template are
//     mutually exclusive single-dimension filters — set at most one.
//   - Sort names the metric rows are ranked by (descending); Limit caps the row
//     count. IncludeTrend/TrendGrain attach a per-bucket rate series to each row
//     where the endpoint supports it.

// EmailStatsSummaryParams filters the summary totals for a window.
type EmailStatsSummaryParams struct {
	// From and To bound the window. Each is a calendar day (YYYY-MM-DD) for a
	// day-grain window (up to 365 days) or an RFC 3339 instant for an hour-grain
	// window (up to 720 hours); both bounds must use the same form. They are
	// strings rather than time.Time because the form the server sees selects the
	// grain, a distinction a time.Time cannot carry.
	From            string
	To              string
	Timezone        string
	Category        string
	SendingDomain   string
	Tag             string
	SendingIp       string
	RecipientDomain string
	Template        string
	// Compare set to "previous_period" also returns the same totals for the
	// immediately preceding window of equal length, plus the change between them.
	Compare string
}

func (p EmailStatsSummaryParams) toWire() *oapi.GetEmailStatsSummaryParams {
	return &oapi.GetEmailStatsSummaryParams{
		From:            statsStr(p.From),
		To:              statsStr(p.To),
		Timezone:        statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:        statsStr(p.Category),
		SendingDomain:   statsStr(p.SendingDomain),
		Tag:             statsStr(p.Tag),
		SendingIp:       statsStr(p.SendingIp),
		RecipientDomain: statsStr(p.RecipientDomain),
		Template:        statsEnum[oapi.EmailStatsTemplateFilter](p.Template),
		Compare:         statsEnum[oapi.GetEmailStatsSummaryParamsCompare](p.Compare),
	}
}

// Summary returns the delivery, engagement, and latency totals for the window,
// optionally with a previous-period comparison.
func (s *EmailStatsService) Summary(ctx context.Context, params EmailStatsSummaryParams, opts ...option.RequestOption) (*EmailStatsSummary, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsSummary(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsSummary
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsDailyParams filters the day-bucketed time series.
type EmailStatsDailyParams struct {
	From            time.Time
	To              time.Time
	Timezone        string
	Category        string
	SendingDomain   string
	Tag             string
	SendingIp       string
	RecipientDomain string
	Template        string
}

func (p EmailStatsDailyParams) toWire() *oapi.GetEmailStatsDailyParams {
	return &oapi.GetEmailStatsDailyParams{
		From:            statsDate(p.From),
		To:              statsDate(p.To),
		Timezone:        statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:        statsStr(p.Category),
		SendingDomain:   statsStr(p.SendingDomain),
		Tag:             statsStr(p.Tag),
		SendingIp:       statsStr(p.SendingIp),
		RecipientDomain: statsStr(p.RecipientDomain),
		Template:        statsEnum[oapi.EmailStatsTemplateFilter](p.Template),
	}
}

// Daily returns one row per calendar day in the window, in chronological order.
func (s *EmailStatsService) Daily(ctx context.Context, params EmailStatsDailyParams, opts ...option.RequestOption) (*EmailStatsResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsDaily(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsHourlyParams filters the hour-bucketed time series. From and To are
// instants, rounded down to the hour; the window may not exceed 720 hours.
type EmailStatsHourlyParams struct {
	From            time.Time
	To              time.Time
	Timezone        string
	Category        string
	SendingDomain   string
	Tag             string
	SendingIp       string
	RecipientDomain string
	Template        string
}

func (p EmailStatsHourlyParams) toWire() *oapi.GetEmailStatsHourlyParams {
	return &oapi.GetEmailStatsHourlyParams{
		From:            statsTime(p.From),
		To:              statsTime(p.To),
		Timezone:        statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:        statsStr(p.Category),
		SendingDomain:   statsStr(p.SendingDomain),
		Tag:             statsStr(p.Tag),
		SendingIp:       statsStr(p.SendingIp),
		RecipientDomain: statsStr(p.RecipientDomain),
		Template:        statsEnum[oapi.EmailStatsTemplateFilter](p.Template),
	}
}

// Hourly returns one row per hour in the window, in chronological order.
func (s *EmailStatsService) Hourly(ctx context.Context, params EmailStatsHourlyParams, opts ...option.RequestOption) (*EmailStatsResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsHourly(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByTagParams ranks a tag breakdown over a day window.
type EmailStatsByTagParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByTagParams) toWire() *oapi.GetEmailStatsByTagParams {
	return &oapi.GetEmailStatsByTagParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByTagParamsTrendGrain](p.TrendGrain),
	}
}

// ByTag ranks statistics per tag (name:value pair) for the window.
func (s *EmailStatsService) ByTag(ctx context.Context, params EmailStatsByTagParams, opts ...option.RequestOption) (*EmailStatsTagsResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByTag(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsTagsResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByCategoryParams ranks the category breakdown over a day window.
// There is no Category filter — this endpoint groups by category.
type EmailStatsByCategoryParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByCategoryParams) toWire() *oapi.GetEmailStatsByCategoryParams {
	return &oapi.GetEmailStatsByCategoryParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByCategoryParamsTrendGrain](p.TrendGrain),
	}
}

// ByCategory ranks statistics per category (transactional, marketing).
func (s *EmailStatsService) ByCategory(ctx context.Context, params EmailStatsByCategoryParams, opts ...option.RequestOption) (*EmailStatsByCategoryResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByCategory(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByCategoryResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsBySendingIpParams ranks the sending-IP breakdown over a day window.
type EmailStatsBySendingIpParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsBySendingIpParams) toWire() *oapi.GetEmailStatsBySendingIpParams {
	return &oapi.GetEmailStatsBySendingIpParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.GetEmailStatsBySendingIpParamsSort](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsBySendingIpParamsTrendGrain](p.TrendGrain),
	}
}

// BySendingIp ranks delivery statistics per sending IP for the window.
func (s *EmailStatsService) BySendingIp(ctx context.Context, params EmailStatsBySendingIpParams, opts ...option.RequestOption) (*EmailStatsBySendingIpResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsBySendingIp(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsBySendingIpResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsBySendingDomainParams ranks the sending-domain breakdown.
type EmailStatsBySendingDomainParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsBySendingDomainParams) toWire() *oapi.GetEmailStatsBySendingDomainParams {
	return &oapi.GetEmailStatsBySendingDomainParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsBySendingDomainParamsTrendGrain](p.TrendGrain),
	}
}

// BySendingDomain ranks statistics per sending domain for the window.
func (s *EmailStatsService) BySendingDomain(ctx context.Context, params EmailStatsBySendingDomainParams, opts ...option.RequestOption) (*EmailStatsBySendingDomainResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsBySendingDomain(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsBySendingDomainResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByRecipientDomainParams ranks the recipient-domain breakdown.
type EmailStatsByRecipientDomainParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByRecipientDomainParams) toWire() *oapi.GetEmailStatsByRecipientDomainParams {
	return &oapi.GetEmailStatsByRecipientDomainParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByRecipientDomainParamsTrendGrain](p.TrendGrain),
	}
}

// ByRecipientDomain ranks statistics per recipient mailbox domain (e.g.
// gmail.com) for the window.
func (s *EmailStatsService) ByRecipientDomain(ctx context.Context, params EmailStatsByRecipientDomainParams, opts ...option.RequestOption) (*EmailStatsByRecipientDomainResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByRecipientDomain(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByRecipientDomainResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByMailboxProviderParams ranks the mailbox-provider breakdown.
type EmailStatsByMailboxProviderParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByMailboxProviderParams) toWire() *oapi.GetEmailStatsByMailboxProviderParams {
	return &oapi.GetEmailStatsByMailboxProviderParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailMailboxProviderSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByMailboxProviderParamsTrendGrain](p.TrendGrain),
	}
}

// ByMailboxProvider ranks post-delivery statistics per mailbox provider (e.g.
// Google, Microsoft) for the window.
func (s *EmailStatsService) ByMailboxProvider(ctx context.Context, params EmailStatsByMailboxProviderParams, opts ...option.RequestOption) (*EmailStatsByMailboxProviderResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByMailboxProvider(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByMailboxProviderResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByMailboxProviderRegionParams ranks the provider-region breakdown.
type EmailStatsByMailboxProviderRegionParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByMailboxProviderRegionParams) toWire() *oapi.GetEmailStatsByMailboxProviderRegionParams {
	return &oapi.GetEmailStatsByMailboxProviderRegionParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailMailboxProviderSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByMailboxProviderRegionParamsTrendGrain](p.TrendGrain),
	}
}

// ByMailboxProviderRegion ranks post-delivery statistics per mailbox-provider
// region for the window.
func (s *EmailStatsService) ByMailboxProviderRegion(ctx context.Context, params EmailStatsByMailboxProviderRegionParams, opts ...option.RequestOption) (*EmailStatsByMailboxProviderRegionResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByMailboxProviderRegion(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByMailboxProviderRegionResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByTemplateParams ranks the template breakdown over a day window.
type EmailStatsByTemplateParams struct {
	From         time.Time
	To           time.Time
	Timezone     string
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByTemplateParams) toWire() *oapi.GetEmailStatsByTemplateParams {
	return &oapi.GetEmailStatsByTemplateParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Timezone:     statsEnum[oapi.StatsTimezone](p.Timezone),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByTemplateParamsTrendGrain](p.TrendGrain),
	}
}

// ByTemplate ranks statistics per template for the window.
func (s *EmailStatsService) ByTemplate(ctx context.Context, params EmailStatsByTemplateParams, opts ...option.RequestOption) (*EmailStatsByTemplateResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByTemplate(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByTemplateResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByLocationParams ranks the geographic breakdown. GroupBy selects the
// granularity: "country" (default), "region", or "city".
type EmailStatsByLocationParams struct {
	From     time.Time
	To       time.Time
	Timezone string
	Category string
	GroupBy  string
	Sort     string
	Limit    int
}

func (p EmailStatsByLocationParams) toWire() *oapi.GetEmailStatsByLocationParams {
	return &oapi.GetEmailStatsByLocationParams{
		From:     statsDate(p.From),
		To:       statsDate(p.To),
		Timezone: statsEnum[oapi.StatsTimezone](p.Timezone),
		Category: statsStr(p.Category),
		GroupBy:  statsEnum[oapi.GetEmailStatsByLocationParamsGroupBy](p.GroupBy),
		Sort:     statsEnum[oapi.EmailEngagementSortMetric](p.Sort),
		Limit:    statsInt(p.Limit),
	}
}

// ByLocation ranks engagement statistics per geographic location for the window.
func (s *EmailStatsService) ByLocation(ctx context.Context, params EmailStatsByLocationParams, opts ...option.RequestOption) (*EmailStatsByLocationResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByLocation(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByLocationResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByClientParams ranks the reading-environment breakdown. GroupBy
// selects the facet: "email_client" (default), "os", or "device_type".
type EmailStatsByClientParams struct {
	From     time.Time
	To       time.Time
	Timezone string
	Category string
	GroupBy  string
	Sort     string
	Limit    int
}

func (p EmailStatsByClientParams) toWire() *oapi.GetEmailStatsByClientParams {
	return &oapi.GetEmailStatsByClientParams{
		From:     statsDate(p.From),
		To:       statsDate(p.To),
		Timezone: statsEnum[oapi.StatsTimezone](p.Timezone),
		Category: statsStr(p.Category),
		GroupBy:  statsEnum[oapi.GetEmailStatsByClientParamsGroupBy](p.GroupBy),
		Sort:     statsEnum[oapi.EmailEngagementSortMetric](p.Sort),
		Limit:    statsInt(p.Limit),
	}
}

// ByClient ranks engagement statistics per reading environment (mail client, OS,
// or device type) for the window.
func (s *EmailStatsService) ByClient(ctx context.Context, params EmailStatsByClientParams, opts ...option.RequestOption) (*EmailStatsByClientResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByClient(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByClientResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByBounceCodeParams ranks the bounce-code breakdown over a day window.
type EmailStatsByBounceCodeParams struct {
	From     time.Time
	To       time.Time
	Timezone string
	Category string
	Sort     string
	Limit    int
}

func (p EmailStatsByBounceCodeParams) toWire() *oapi.GetEmailStatsByBounceCodeParams {
	return &oapi.GetEmailStatsByBounceCodeParams{
		From:     statsDate(p.From),
		To:       statsDate(p.To),
		Timezone: statsEnum[oapi.StatsTimezone](p.Timezone),
		Category: statsStr(p.Category),
		Sort:     statsEnum[oapi.GetEmailStatsByBounceCodeParamsSort](p.Sort),
		Limit:    statsInt(p.Limit),
	}
}

// ByBounceCode ranks bounce counts per SMTP error code for the window.
func (s *EmailStatsService) ByBounceCode(ctx context.Context, params EmailStatsByBounceCodeParams, opts ...option.RequestOption) (*EmailStatsByBounceCodeResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByBounceCode(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByBounceCodeResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByComplaintTypeParams ranks the complaint-type breakdown.
type EmailStatsByComplaintTypeParams struct {
	From     time.Time
	To       time.Time
	Timezone string
	Category string
	Sort     string
	Limit    int
}

func (p EmailStatsByComplaintTypeParams) toWire() *oapi.GetEmailStatsByComplaintTypeParams {
	return &oapi.GetEmailStatsByComplaintTypeParams{
		From:     statsDate(p.From),
		To:       statsDate(p.To),
		Timezone: statsEnum[oapi.StatsTimezone](p.Timezone),
		Category: statsStr(p.Category),
		Sort:     statsEnum[oapi.GetEmailStatsByComplaintTypeParamsSort](p.Sort),
		Limit:    statsInt(p.Limit),
	}
}

// ByComplaintType ranks complaint counts per feedback-loop complaint type for
// the window.
func (s *EmailStatsService) ByComplaintType(ctx context.Context, params EmailStatsByComplaintTypeParams, opts ...option.RequestOption) (*EmailStatsByComplaintTypeResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByComplaintType(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByComplaintTypeResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EmailStatsByBroadcastParams ranks the broadcast breakdown over a day window.
type EmailStatsByBroadcastParams struct {
	From         time.Time
	To           time.Time
	Category     string
	Sort         string
	Limit        int
	IncludeTrend bool
	TrendGrain   string
}

func (p EmailStatsByBroadcastParams) toWire() *oapi.GetEmailStatsByBroadcastParams {
	return &oapi.GetEmailStatsByBroadcastParams{
		From:         statsDate(p.From),
		To:           statsDate(p.To),
		Category:     statsStr(p.Category),
		Sort:         statsEnum[oapi.EmailStatsSortMetric](p.Sort),
		Limit:        statsInt(p.Limit),
		IncludeTrend: statsBool(p.IncludeTrend),
		TrendGrain:   statsEnum[oapi.GetEmailStatsByBroadcastParamsTrendGrain](p.TrendGrain),
	}
}

// ByBroadcast ranks statistics per broadcast for the window.
func (s *EmailStatsService) ByBroadcast(ctx context.Context, params EmailStatsByBroadcastParams, opts ...option.RequestOption) (*EmailStatsByBroadcastResponse, error) {
	body, err := s.get(ctx, opts, func(ctx context.Context, cfg requestConfig) (*http.Response, error) {
		return s.client.oapi.GetEmailStatsByBroadcast(ctx, params.toWire(), cfg...)
	})
	if err != nil {
		return nil, err
	}
	var out EmailStatsByBroadcastResponse
	if err := decodeBody(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// requestConfig is the per-request editor slice a stats call hands to the
// generated low-level client.
type requestConfig = []oapi.RequestEditorFn

// get runs a read through the shared request core: it resolves per-call
// options, then executes call (never a mutation, so no idempotency key) and
// returns the decoded response body.
func (s *EmailStatsService) get(ctx context.Context, opts []option.RequestOption, call func(context.Context, requestConfig) (*http.Response, error)) ([]byte, error) {
	cfg, err := s.client.resolve(opts)
	if err != nil {
		return nil, err
	}
	return cfg.Execute(ctx, false, func(ctx context.Context, _ string) (*http.Response, error) {
		return call(ctx, s.client.callEditors(cfg))
	})
}

// statsDate renders a calendar-day query param; a zero time is omitted.
func statsDate(t time.Time) *openapi_types.Date {
	if t.IsZero() {
		return nil
	}
	return &openapi_types.Date{Time: t}
}

// statsTime renders an instant query param; a zero time is omitted.
func statsTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func statsStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func statsInt(n int) *int {
	if n <= 0 {
		return nil
	}
	return &n
}

func statsBool(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

// statsEnum casts a caller-supplied string to a named wire enum (or string
// alias) query param; an empty string is omitted.
func statsEnum[T ~string](s string) *T {
	if s == "" {
		return nil
	}
	v := T(s)
	return &v
}
