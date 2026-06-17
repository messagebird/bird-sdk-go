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

// EmailTag is a structured {Name, Value} label.
type EmailTag = oapi.EmailTag

// EmailAttachment is a file attachment on a send.
type EmailAttachment = oapi.EmailAttachment

// EmailStatus is a message's aggregate delivery status.
type EmailStatus = oapi.EmailMessageStatus

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

// WebhookEventType is a webhook event's discriminant.
type WebhookEventType = oapi.WebhookEventType

const (
	EventTypeDomainFailed            WebhookEventType = "domain.failed"
	EventTypeDomainVerified          WebhookEventType = "domain.verified"
	EventTypeEmailAccepted           WebhookEventType = "email.accepted"
	EventTypeEmailBounced            WebhookEventType = "email.bounced"
	EventTypeEmailClicked            WebhookEventType = "email.clicked"
	EventTypeEmailComplained         WebhookEventType = "email.complained"
	EventTypeEmailDeferred           WebhookEventType = "email.deferred"
	EventTypeEmailDelivered          WebhookEventType = "email.delivered"
	EventTypeEmailListUnsubscribed   WebhookEventType = "email.list_unsubscribed"
	EventTypeEmailOpened             WebhookEventType = "email.opened"
	EventTypeEmailOutOfBandBounce    WebhookEventType = "email.out_of_band_bounce"
	EventTypeEmailProcessed          WebhookEventType = "email.processed"
	EventTypeEmailReceived           WebhookEventType = "email.received"
	EventTypeEmailRejected           WebhookEventType = "email.rejected"
	EventTypeEmailSuppressionCreated WebhookEventType = "email_suppression.created"
	EventTypeEmailUnsubscribed       WebhookEventType = "email.unsubscribed"
)

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
