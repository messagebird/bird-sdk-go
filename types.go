package bird

import (
	"github.com/messagebird/bird-sdk-go/internal/oapi"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Ptr returns a pointer to v, for setting optional pointer fields inline. Bool,
// String, Int, and Email are typed shorthands for the common cases:
//
//	bird.EmailSendParams{TrackOpens: bird.Bool(false)}
func Ptr[T any](v T) *T { return &v }

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// String returns a pointer to v.
func String(v string) *string { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Email returns a pointer to v as an email-typed field. A schema's `format: email`
// property is its own type on the wire, so bird.String does not fit one and Ptr
// needs the conversion spelled out:
//
//	bird.VerificationTo{EmailAddress: bird.Email("user@example.com")}
func Email(v string) *openapi_types.Email {
	e := openapi_types.Email(v)
	return &e
}

// Nullable is a nullable/clearable request-param field. It carries one of three
// states: a value, an explicit JSON null (clears the field), or unspecified (the
// zero value — omitted, leaving the field unchanged). Only request params use it;
// response fields stay plain pointers. Build it with Value or Null:
//
//	bird.AudienceUpdateParams{Description: bird.Null[string]()}   // clear
//	bird.AudienceUpdateParams{Description: bird.Value("Q4 leads")} // set
type Nullable[T any] = nullable.Nullable[T]

// Value sets a Nullable request field to send v.
func Value[T any](v T) Nullable[T] { return nullable.NewNullableWithValue(v) }

// Null sets a Nullable request field to send an explicit JSON null, clearing it.
func Null[T any]() Nullable[T] { return nullable.NewNullNullable[T]() }

// Public type aliases — the semver-locked names for the SDK's response and
// enum types. Typed IDs are plain strings on the wire, so they need no alias.

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

// Enum vocabularies the read filters expose. Each is a named type, so a params
// field carries it rather than a bare string.
type (
	// TemplateScope distinguishes Bird's built-in templates from a workspace's own.
	TemplateScope = oapi.TemplateScope
	// SMSMessageCategory is an SMS's content classification.
	SMSMessageCategory = oapi.SMSMessageCategory
	// EmailStatsSortMetric is the metric an email-stats breakdown sorts by.
	EmailStatsSortMetric = oapi.EmailStatsSortMetric
	// EmailEngagementSortMetric is the engagement metric a breakdown sorts by.
	EmailEngagementSortMetric = oapi.EmailEngagementSortMetric
	// EmailMailboxProviderSortMetric is the metric a mailbox-provider breakdown
	// sorts by.
	EmailMailboxProviderSortMetric = oapi.EmailMailboxProviderSortMetric
	// StatsTrendGrain is the bucket grain of a stats trend series.
	StatsTrendGrain = oapi.StatsTrendGrain
	// MessageDirection is whether a message was sent or received.
	MessageDirection = oapi.MessageDirection
	// EmailMessageStatus is an alias of EmailStatus, used by the read filters.
	EmailMessageStatus = oapi.EmailMessageStatus
	// EmailMessageCategory is an alias of Category, used by the read filters.
	EmailMessageCategory = oapi.EmailMessageCategory
)

// Email statistics responses, returned by the Client.Email.Stats methods. Each
// is the read-side body for one breakdown.
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
	EmailStatsBySendingIPResponse             = oapi.EmailStatsBySendingIpResponse
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
// The write-side *Config aliases accompany DomainCreateParams / DomainUpdateParams.
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

// Verification is a verification's current state (id, status, channel plan);
// VerificationCheckResult is a check outcome plus the verification's state.
type (
	Verification            = oapi.Verification
	VerificationCheckResult = oapi.VerificationCheckResult
)

// Realtime read and publish results. RealtimePublishResult and
// RealtimeBatchPublishResult carry per-channel counts only when the call asked
// for them via Include. RealtimeChannelsList is the app's occupied channels
// (unpaginated); RealtimeChannelInfo is one channel's state;
// RealtimeChannelMembers is the members present on a presence channel.
type (
	RealtimePublishResult          = oapi.RealtimePublishResult
	RealtimeBatchPublishResult     = oapi.RealtimeBatchPublishResult
	RealtimeBatchPublishResultItem = oapi.RealtimeBatchPublishResultItem
	RealtimeChannelsList           = oapi.RealtimeChannelsList
	RealtimeChannelListItem        = oapi.RealtimeChannelListItem
	RealtimeChannelInfo            = oapi.RealtimeChannelInfo
	RealtimeChannelMembers         = oapi.RealtimeChannelMembers
	RealtimeChannelMember          = oapi.RealtimeChannelMember
)

// RealtimeChannelInclude names a per-channel attribute to return alongside a
// publish or channel read.
type RealtimeChannelInclude = oapi.RealtimeChannelInclude

const (
	// RealtimeIncludeMemberCount is presence-channels only.
	RealtimeIncludeMemberCount RealtimeChannelInclude = "member_count"
	// RealtimeIncludeConnectionCount requires the app's connection-counting flag.
	RealtimeIncludeConnectionCount RealtimeChannelInclude = "connection_count"
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

// WebhookEventType is a webhook event's discriminant. It is an open string: the
// known values are the EventType* constants, and an event type added by a newer
// server flows through Unwrap as a plain string.
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

// Agent mailbox types.
type (
	// Mailbox is a durable inbox on inbox.ai or a custom domain.
	Mailbox = oapi.Mailbox
	// MailboxList is one page of mailboxes plus its pagination cursors.
	MailboxList = oapi.MailboxList
	// MailboxStatsResponse is the stats time series for a mailbox.
	MailboxStatsResponse = oapi.MailboxStatsResponse
	// EmailMailboxLabelList is the list of labels available in a mailbox.
	EmailMailboxLabelList = oapi.EmailMailboxLabelList

	// ReceiveRule is a per-sender allow or block rule on a mailbox.
	ReceiveRule = oapi.ReceiveRule
	// ReceiveRuleList is one page of receive rules.
	ReceiveRuleList = oapi.ReceiveRuleList

	// EmailThread is a conversation: a group of messages on the same topic.
	EmailThread = oapi.EmailThread
	// EmailThreadList is one page of threads plus its pagination cursors.
	EmailThreadList = oapi.EmailThreadList

	// EmailThreadMessage is a single message in a conversation.
	EmailThreadMessage = oapi.EmailThreadMessage
	// EmailThreadMessageList is one page of messages.
	EmailThreadMessageList = oapi.EmailThreadMessageList
	// EmailThreadMessageBody is the parsed HTML and plain-text body.
	EmailThreadMessageBody = oapi.EmailThreadMessageBody
	// EmailThreadMessageAttachmentList is the attachment manifest.
	EmailThreadMessageAttachmentList = oapi.EmailThreadMessageAttachmentList
)
