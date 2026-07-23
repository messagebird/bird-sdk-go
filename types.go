package bird

import "github.com/messagebird/bird-sdk-go/internal/oapi"

// Ptr returns a pointer to v, for setting optional pointer fields inline. Bool,
// String, and Int are typed shorthands for the common cases:
//
//	bird.EmailSendParams{TrackOpens: bird.Bool(false)}
func Ptr[T any](v T) *T { return &v }

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// String returns a pointer to v.
func String(v string) *string { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Public aliases for the curated surface. The generated types live in
// internal/oapi; these names are the semver-locked commitment. Typed IDs are
// plain strings on the wire, so they need no alias.

// EmailMessage is a sent message with aggregate delivery status.
type EmailMessage = oapi.EmailMessage

// EmailMessageList is one page of messages plus its pagination cursors.
type EmailMessageList = oapi.EmailMessageList

// EmailBatch is the result of a batch send: one item per submitted message, in
// submission order.
type EmailBatch = oapi.EmailMessageBatchResponse

// EmailBatchItem is a single message's entry in a batch send result.
type EmailBatchItem = oapi.EmailMessageBatchItem

// EmailTag is a structured {Name, Value} label.
type EmailTag = oapi.Tag

// EmailAttachment is a file attachment on a send.
type EmailAttachment = oapi.EmailAttachment

// EmailStatus is a message's aggregate delivery status.
type EmailStatus = oapi.EmailMessageStatus

// Email statistics responses, returned by the Client.Email.Stats methods. Each
// is the read-side body for one breakdown; the shared metric shapes (delivery,
// engagement, latency, and the per-row points) live in internal/oapi and are
// reachable through these aliases.
type (
	// EmailStatsSummary is the delivery/engagement/latency totals for a window,
	// optionally with a previous-period comparison. Returned by Stats.Summary.
	EmailStatsSummary = oapi.EmailStatsSummary
	// EmailStatsResponse is a time series of per-bucket points. Returned by
	// Stats.Daily and Stats.Hourly.
	EmailStatsResponse = oapi.EmailStatsResponse
	// EmailStatsTagsResponse is the ranked tag breakdown. Returned by Stats.ByTag.
	EmailStatsTagsResponse                    = oapi.EmailStatsTagsResponse
	EmailStatsByCategoryResponse              = oapi.EmailStatsByCategoryResponse
	EmailStatsBySendingIpResponse             = oapi.EmailStatsBySendingIpResponse
	EmailStatsBySendingDomainResponse         = oapi.EmailStatsBySendingDomainResponse
	EmailStatsByRecipientDomainResponse       = oapi.EmailStatsByRecipientDomainResponse
	EmailStatsByMailboxProviderResponse       = oapi.EmailStatsByMailboxProviderResponse
	EmailStatsByMailboxProviderRegionResponse = oapi.EmailStatsByMailboxProviderRegionResponse
	EmailStatsByTemplateResponse              = oapi.EmailStatsByTemplateResponse
	EmailStatsByLocationResponse              = oapi.EmailStatsByLocationResponse
	EmailStatsByClientResponse                = oapi.EmailStatsByClientResponse
	EmailStatsByBounceCodeResponse            = oapi.EmailStatsByBounceCodeResponse
	EmailStatsByComplaintTypeResponse         = oapi.EmailStatsByComplaintTypeResponse
	EmailStatsByBroadcastResponse             = oapi.EmailStatsByBroadcastResponse
)

// SMSTemplate is an SMS template with its body, variables, and available
// languages; SMSTemplateList is the (unpaginated) set of templates available to
// the workspace.
type (
	SMSTemplate     = oapi.SMSTemplate
	SMSTemplateList = oapi.SMSTemplateList
)

// SMSMessage is a sent or received SMS with its status, segment breakdown, and
// cost; SMSMessageList is a page of messages; SMSBatch is a batch-send result.
type (
	SMSMessage     = oapi.SMSMessage
	SMSMessageList = oapi.SMSMessageList
	SMSBatch       = oapi.SMSMessageBatchResponse
)

// SMSTag is a structured {name, value} label on an SMS send.
type SMSTag = oapi.Tag

// SMSStatus is a message's delivery status.
type SMSStatus = oapi.SMSMessageStatus

// Contact is a workspace contact; ContactList is a page of contacts;
// ContactUpsertResult is the result of a bulk upsert, with one
// ContactUpsertResultItem per submitted contact in submission order.
type (
	Contact                 = oapi.Contact
	ContactList             = oapi.ContactList
	ContactUpsertResult     = oapi.ContactUpsertResult
	ContactUpsertResultItem = oapi.ContactUpsertResultItem
)

// Audience is a static audience of contacts; AudienceList is a page of
// audiences. AudienceMember pairs a contact with the time it joined;
// AudienceMemberList is a page of members.
type (
	Audience           = oapi.Audience
	AudienceList       = oapi.AudienceList
	AudienceMember     = oapi.AudienceMember
	AudienceMemberList = oapi.AudienceMemberList
)

// ContactProperty is a custom contact property definition; ContactPropertyList
// is a page of properties.
type (
	ContactProperty     = oapi.ContactProperty
	ContactPropertyList = oapi.ContactPropertyList
)

// Domain is a sending domain with its DNS records and per-capability status;
// DomainList is a page of domains. DNSRecord is one required DNS record and its
// verification state; DomainDKIM is the domain's active DKIM signing
// configuration; DomainCapabilities is the per-capability readiness breakdown.
// These are the read-side types; the write-side configs are the *Config structs
// in domains.go.
type (
	Domain             = oapi.Domain
	DomainList         = oapi.DomainList
	DNSRecord          = oapi.DNSRecord
	DomainDKIM         = oapi.DomainDKIM
	DomainCapabilities = oapi.DomainCapabilities
	DomainCapability   = oapi.DomainCapability
	DomainStatus       = oapi.DomainStatus
)

// WhatsAppMessage is a sent or received WhatsApp message; WhatsAppMessageList
// is a page of messages.
type (
	WhatsAppMessage     = oapi.WhatsAppMessage
	WhatsAppMessageList = oapi.WhatsAppMessageList
)

// WhatsAppMessageStatus is a message's delivery status.
type WhatsAppMessageStatus = oapi.WhatsAppMessageStatus

// WhatsAppEvent is a single lifecycle event on a message's timeline;
// WhatsAppEventList is the (unpaginated) timeline for one message.
type (
	WhatsAppEvent     = oapi.WhatsAppEvent
	WhatsAppEventList = oapi.WhatsAppEventList
)

// WhatsAppMessageTemplateComponent is a filled-in template component — supplied
// on a template send and echoed back on the sent message.
// WhatsAppMessageTemplateComponentParameter is one of its placeholder values.
type (
	WhatsAppMessageTemplateComponent          = oapi.WhatsAppMessageTemplateComponent
	WhatsAppMessageTemplateComponentParameter = oapi.WhatsAppMessageTemplateComponentParameter
)

// WhatsAppTemplate is a template available to the workspace; WhatsAppTemplateList
// is the (unpaginated) set of templates.
type (
	WhatsAppTemplate     = oapi.WhatsAppTemplate
	WhatsAppTemplateList = oapi.WhatsAppTemplateList
)

// Verification is a verification's current state (id, status, channel plan);
// VerificationCheckResult is a check outcome plus the verification's state.
type (
	Verification            = oapi.Verification
	VerificationCheckResult = oapi.VerificationCheckResult
)

const (
	EmailStatusAccepted       EmailStatus = "accepted"
	EmailStatusProcessed      EmailStatus = "processed"
	EmailStatusDelivered      EmailStatus = "delivered"
	EmailStatusDeferred       EmailStatus = "deferred"
	EmailStatusBounced        EmailStatus = "bounced"
	EmailStatusComplained     EmailStatus = "complained"
	EmailStatusRejected       EmailStatus = "rejected"
	EmailStatusPartialFailure EmailStatus = "partial_failure"
)

// Category classifies a send's suppression policy.
type Category = oapi.EmailMessageCategory

const (
	CategoryTransactional Category = "transactional"
	CategoryMarketing     Category = "marketing"
)

// WebhookEventType is a webhook event's discriminant. It is an open string:
// the known values are the EventType* constants in eventtypes.gen.go, and an
// event type added by a newer server flows through Unwrap as a plain string.
type WebhookEventType = oapi.WebhookEventType

// Webhook event payloads, returned by Event.AsAny. Type-switch on these.
type (
	DomainFailedEvent            = oapi.EventDomainFailed
	DomainVerifiedEvent          = oapi.EventDomainVerified
	EmailAcceptedEvent           = oapi.EventEmailAccepted
	EmailBouncedEvent            = oapi.EventEmailBounced
	EmailClickedEvent            = oapi.EventEmailClicked
	EmailComplainedEvent         = oapi.EventEmailComplained
	EmailDeferredEvent           = oapi.EventEmailDeferred
	EmailDeliveredEvent          = oapi.EventEmailDelivered
	EmailListUnsubscribedEvent   = oapi.EventEmailListUnsubscribed
	EmailOpenedEvent             = oapi.EventEmailOpened
	EmailOutOfBandBounceEvent    = oapi.EventEmailOutOfBandBounce
	EmailProcessedEvent          = oapi.EventEmailProcessed
	EmailReceivedEvent           = oapi.EventEmailReceived
	EmailRejectedEvent           = oapi.EventEmailRejected
	EmailSuppressionCreatedEvent = oapi.EventEmailSuppressionCreated
	EmailUnsubscribedEvent       = oapi.EventEmailUnsubscribed
)
