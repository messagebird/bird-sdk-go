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

// EmailTemplate is a template with its current draft content and version
// pointers; EmailTemplateSummary is the list-item form; EmailTemplateVersion is
// a single version; EmailTemplateList is a page of summaries;
// EmailTemplateVersionList is the (unpaginated) set of a template's versions.
type (
	EmailTemplate            = oapi.EmailTemplate
	EmailTemplateSummary     = oapi.EmailTemplateSummary
	EmailTemplateVersion     = oapi.EmailTemplateVersion
	EmailTemplateList        = oapi.EmailTemplateList
	EmailTemplateVersionList = oapi.EmailTemplateVersionList
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
