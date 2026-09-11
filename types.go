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
//	bird.VerificationTo{Email: bird.Email("user@example.com")}
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

// Broadcast reads, returned by the Client.Broadcasts methods.
type (
	// EmailBroadcast is one broadcast with its audience reference and counters.
	EmailBroadcast = oapi.EmailBroadcast
	// EmailBroadcastList is one page of broadcasts plus its pagination cursors.
	EmailBroadcastList = oapi.EmailBroadcastList
	// EmailBroadcastCounts is how many contacts a broadcast's audience holds,
	// how many of those have an address, and how many of those addressable ones
	// are not suppressed. They describe the audience, not delivery, and sending
	// does not change them.
	EmailBroadcastCounts = oapi.EmailBroadcastCounts
	// EmailBroadcastClickedLink is one link in a broadcast, with how many
	// recipients clicked it.
	EmailBroadcastClickedLink = oapi.EmailBroadcastClickedLink
	// EmailBroadcastClickedLinkList is the links a broadcast's recipients
	// clicked, with a click count each.
	EmailBroadcastClickedLinkList = oapi.EmailBroadcastClickedLinkList
	// EmailBroadcastSendQuota is how much of a broadcast the organization's
	// email send allowance covers, read before sending it.
	EmailBroadcastSendQuota = oapi.EmailBroadcastSendQuota
	// EmailBroadcastStatus is where a broadcast stands in its lifecycle, reported
	// as EmailBroadcastCounts.Status and taken by the broadcast list filter.
	EmailBroadcastStatus = oapi.EmailBroadcastStatus
	// EmailRecipient is one address a broadcast sent to, with its own state.
	EmailRecipient = oapi.EmailRecipient
	// EmailRecipientList is one page of recipients plus its pagination cursors.
	EmailRecipientList = oapi.EmailRecipientList
	// EmailEvent is one entry in a delivery timeline.
	EmailEvent = oapi.EmailEvent
	// EmailEventList is one page of events plus its pagination cursors.
	EmailEventList = oapi.EmailEventList
	// EmailBroadcastCategory is a broadcast's suppression policy, as every read
	// reports it. It is a separate type from the shared Category, so passing a
	// read value where a Category is wanted converts: Category(b.Category).
	EmailBroadcastCategory = oapi.EmailBroadcastCategory
	// EmailBroadcastFailureReason is why a failed broadcast stopped.
	EmailBroadcastFailureReason = oapi.EmailBroadcastFailureReason
	// EmailSendAllowanceWindow is the period a send allowance is measured over,
	// reported as EmailBroadcastSendQuota.LimitedBy.
	EmailSendAllowanceWindow = oapi.EmailSendAllowanceWindow
	// EmailRecipientStatus is where one recipient of a broadcast has got to.
	EmailRecipientStatus = oapi.EmailRecipientStatus
	// EmailEventBounceType is how a receiving server refused a message.
	EmailEventBounceType = oapi.EmailEventBounceType
	// EmailRecipientBounceType is how a receiving server refused this recipient's
	// copy. Distinct from EmailEventBounceType: a recipient carries its own.
	EmailRecipientBounceType = oapi.EmailRecipientBounceType
	// EmailEventRejectionReason is why we refused an event's send before trying.
	EmailEventRejectionReason = oapi.EmailEventRejectionReason
	// EmailRecipientRejectionReason is why we refused this recipient before trying.
	EmailRecipientRejectionReason = oapi.EmailRecipientRejectionReason
	// RecipientRole is which address field a recipient was named in.
	RecipientRole = oapi.RecipientRole
)

// Enum vocabularies the read filters expose. Each is a named type, so a params
// field carries it rather than a bare string.
type (
	// TemplateScope distinguishes Bird's built-in templates from a workspace's own.
	TemplateScope = oapi.TemplateScope
	// TemplateStatus is where a template stands as a whole, on every channel.
	TemplateStatus = oapi.TemplateStatus
	// EmailTemplateThemeFilter is the closed set of themes the template list filter
	// takes. EmailTemplateTheme, the value a template reports back, is open, and the
	// exported EmailTemplateTheme* constants are its type, not this one -- a filter
	// takes a string literal, as every other closed enum here does.
	EmailTemplateThemeFilter = oapi.EmailTemplateThemeFilter
	// EmailTemplateCategory is whether a template is transactional or marketing.
	EmailTemplateCategory = oapi.EmailTemplateCategory
	// EmailTemplateSourceWrite is the closed set of authoring formats a template
	// can be created in. EmailTemplateSource, the value a template reports back,
	// is open, because a format Bird adds later has to parse on a client that
	// shipped before it.
	EmailTemplateSourceWrite = oapi.EmailTemplateSourceWrite
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
	// ContactIdentifierFilter is which identifier a contact has on file, used by the read filters.
	ContactIdentifierFilter = oapi.ContactIdentifierFilter
	// SMSStatsSortMetric is the metric an SMS-stats breakdown sorts by, for the
	// breakdowns whose rows carry rates as well as counts.
	SMSStatsSortMetric = oapi.SMSStatsSortMetric
	// SMSStatsLifecycleSortMetric is the reduced sort vocabulary of a breakdown
	// whose rows carry lifecycle counts only.
	SMSStatsLifecycleSortMetric = oapi.SMSStatsLifecycleSortMetric
	// StatsComparePeriod asks a stats summary for the preceding window too. Not
	// SMS-specific: every channel's summary read takes the same one value, so the
	// four still declaring it inline can $ref this instead.
	StatsComparePeriod = oapi.StatsComparePeriod
	// SMSSuppressionReasonFilter is why a suppression exists, used by the read filters.
	SMSSuppressionReasonFilter = oapi.SMSSuppressionReasonFilter
	// SMSKeywordRuleScope distinguishes Bird's default keyword rules from a
	// workspace's own.
	SMSKeywordRuleScope = oapi.SMSKeywordRuleScope
)

// Hand-written because closing the enum took TemplateStatus out of the
// open-enum generator's reach, and these five were already published.
const (
	TemplateStatusDraft    = oapi.TemplateStatusDraft
	TemplateStatusPending  = oapi.TemplateStatusPending
	TemplateStatusActive   = oapi.TemplateStatusActive
	TemplateStatusRejected = oapi.TemplateStatusRejected
	TemplateStatusInactive = oapi.TemplateStatusInactive
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

// Email template reads, returned by the Client.Email.Templates methods.
type (
	// EmailTemplate is one template with its draft and published versions.
	EmailTemplate = oapi.EmailTemplate
	// EmailTemplateSummary is a template's list row.
	EmailTemplateSummary = oapi.EmailTemplateSummary
	// EmailTemplateList is one page of templates plus its pagination cursors.
	EmailTemplateList = oapi.EmailTemplateList
	// EmailTemplatePreview is a template rendered with sample values, with the
	// client-compatibility findings the render turned up.
	EmailTemplatePreview = oapi.EmailTemplatePreview
	// EmailTemplateBroadcastSummary is one broadcast blocking a template delete.
	EmailTemplateBroadcastSummary = oapi.EmailTemplateBroadcastSummary
	// EmailTemplateBroadcastList is one page of those broadcasts plus its
	// pagination cursors.
	EmailTemplateBroadcastList = oapi.EmailTemplateBroadcastList
	// EmailTemplateVersion is one version of a template with its per-language
	// content.
	EmailTemplateVersion = oapi.EmailTemplateVersion
	// EmailTemplateVersionSummary is a version's list row.
	EmailTemplateVersionSummary = oapi.EmailTemplateVersionSummary
	// EmailTemplateVersionList is one page of versions plus its pagination
	// cursors.
	EmailTemplateVersionList = oapi.EmailTemplateVersionList
	// EmailTemplateSubmitResult is the outcome of submitting a draft: the frozen
	// version, or the problems that stopped it freezing.
	EmailTemplateSubmitResult = oapi.EmailTemplateSubmitResult
	// EmailTemplateLanguage is one language's subject and body on a version.
	EmailTemplateLanguage = oapi.EmailTemplateLanguage
	// EmailTemplateLanguageList is every language a version carries.
	EmailTemplateLanguageList = oapi.EmailTemplateLanguageList
	// EmailTemplateLanguageSaved is a written language and its new revision,
	// with the same advisory client-compatibility report the language read
	// carries.
	EmailTemplateLanguageSaved = oapi.EmailTemplateLanguageSaved
	// EmailTemplateLanguageContent is one language's subject and bodies, the
	// value side of the map Create takes.
	EmailTemplateLanguageContent = oapi.EmailTemplateLanguageContent
)

// SMSTemplate is an SMS template with its body, variables, and available
// languages; SMSTemplateList is the (unpaginated) set of templates available to
// the workspace.

// SMS statistics responses, returned by the Client.Sms.Stats methods. Each is
// the read-side body for one breakdown; the Inbound set counts messages the
// workspace's own numbers received rather than what it sent.
type (
	// SMSStatsSummary is the delivery and latency totals for a window,
	// optionally with a previous-period comparison. Returned by Stats.Summary.
	SMSStatsSummary = oapi.SMSStatsSummary
	// SMSStatsResponse is a time series of per-bucket points. Returned by
	// Stats.Daily and Stats.Hourly.
	SMSStatsResponse             = oapi.SMSStatsResponse
	SMSStatsByCountryResponse    = oapi.SMSStatsByCountryResponse
	SMSStatsByCarrierResponse    = oapi.SMSStatsByCarrierResponse
	SMSStatsByCategoryResponse   = oapi.SMSStatsByCategoryResponse
	SMSStatsByOriginatorResponse = oapi.SMSStatsByOriginatorResponse
	SMSStatsByStatusResponse     = oapi.SMSStatsByStatusResponse
	SMSStatsByErrorCodeResponse  = oapi.SMSStatsByErrorCodeResponse
	SMSStatsByTagResponse        = oapi.SMSStatsByTagResponse

	SMSInboundStatsSummaryResponse    = oapi.SMSInboundStatsSummaryResponse
	SMSInboundStatsResponse           = oapi.SMSInboundStatsResponse
	SMSInboundStatsByCountryResponse  = oapi.SMSInboundStatsByCountryResponse
	SMSInboundStatsByOperatorResponse = oapi.SMSInboundStatsByOperatorResponse
	SMSInboundStatsByNumberResponse   = oapi.SMSInboundStatsByNumberResponse
)

// SMSEventList is the lifecycle timeline of one SMS, oldest first. Returned by
// Sms.ListEvents.
type SMSEventList = oapi.SMSEventList

// SMSKeywordRule is what one keyword does when a subscriber texts it, Bird's or
// the workspace's own; SMSKeywordRuleList is the (unpaginated) set of them.
type (
	SMSKeywordRule     = oapi.SMSKeywordRule
	SMSKeywordRuleList = oapi.SMSKeywordRuleList
)

// SMSSuppression is one sender-and-subscriber pair Bird will not deliver to;
// SMSSuppressionList is a page of them.
type (
	SMSSuppression     = oapi.SMSSuppression
	SMSSuppressionList = oapi.SMSSuppressionList
)

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

// Preference is one recorded consent grant or opt-out; PreferenceList is a
// page of them.
type (
	Preference     = oapi.Preference
	PreferenceList = oapi.PreferenceList
)

// PreferenceStatus is what a statement says: granted records consent to
// receive messages, revoked records an opt-out.
type PreferenceStatus = oapi.PreferenceStatus

const (
	PreferenceStatusGranted PreferenceStatus = "granted"
	PreferenceStatusRevoked PreferenceStatus = "revoked"
)

// PreferenceCoverage is how much traffic a statement covers:
// non_transactional keeps transactional messages such as receipts and
// verification codes flowing; all covers every message.
type PreferenceCoverage = oapi.PreferenceCoverage

const (
	PreferenceCoverageAll              PreferenceCoverage = "all"
	PreferenceCoverageNonTransactional PreferenceCoverage = "non_transactional"
)

// PreferenceWriteResult is the outcome of a preference create or delete.
// Applied true means the write took effect; applied false means it was
// refused as older than the key's current statement, and Preference then
// carries the statement that survived.
type PreferenceWriteResult = oapi.PreferenceWriteResult

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
	Domain     = oapi.Domain
	DomainList = oapi.DomainList

	// Number is a number the workspace holds; NumberList is a page of them.
	// AvailableNumber is one on sale and AvailableNumberList a page of those.
	// NumbersOrder is a purchase and NumbersOrderList a page of purchases.
	Number              = oapi.Number
	NumberList          = oapi.NumberList
	AvailableNumber     = oapi.AvailableNumber
	AvailableNumberList = oapi.AvailableNumberList
	NumbersOrder        = oapi.NumbersOrder
	NumbersOrderList    = oapi.NumbersOrderList
	DNSRecord           = oapi.DNSRecord
	DomainDKIM          = oapi.DomainDKIM
	DomainCapabilities  = oapi.DomainCapabilities
	DomainCapability    = oapi.DomainCapability
	DomainStatus        = oapi.DomainStatus
)

// WhatsAppMessage is a sent or received WhatsApp message; WhatsAppMessageList
// is a page of messages.
type (
	WhatsAppMessage     = oapi.WhatsAppMessage
	WhatsAppMessageList = oapi.WhatsAppMessageList
)

// WhatsAppMessageStatus is a message's delivery status.
type WhatsAppMessageStatus = oapi.WhatsAppMessageStatus

// WhatsAppReadReceipt is the acknowledgement Bird accepted for an inbound
// message. Nothing reports back afterwards: WhatsApp publishes no callback for
// one.
type WhatsAppReadReceipt = oapi.WhatsAppReadReceipt

// WhatsAppReaction is a reaction standing on a message right now, one per
// sender; WhatsAppReactionAccepted is the reaction a place request accepted,
// which names the log entry it created.
type (
	WhatsAppReaction         = oapi.WhatsAppReaction
	WhatsAppReactionAccepted = oapi.WhatsAppReactionAccepted
)

// WhatsAppReactionEvent is one change to a message's reactions — an emoji
// placed, replaced, or taken back; WhatsAppReactionEventList is a page of them.
type (
	WhatsAppReactionEvent     = oapi.WhatsAppReactionEvent
	WhatsAppReactionEventList = oapi.WhatsAppReactionEventList
)

// WhatsAppEvent is a single lifecycle event on a message's timeline;
// WhatsAppEventList is the (unpaginated) timeline for one message.
type (
	WhatsAppEvent     = oapi.WhatsAppEvent
	WhatsAppEventList = oapi.WhatsAppEventList
)

// The template registry, read at three levels: a template is the handle a send
// names, a version is one immutable submission of it, and a language is one
// language's content within a version. The Summary forms carry no content.
type (
	WhatsAppTemplate                = oapi.WhatsAppTemplate
	WhatsAppTemplateList            = oapi.WhatsAppTemplateList
	WhatsAppTemplateVersion         = oapi.WhatsAppTemplateVersion
	WhatsAppTemplateVersionSummary  = oapi.WhatsAppTemplateVersionSummary
	WhatsAppTemplateVersionList     = oapi.WhatsAppTemplateVersionList
	WhatsAppTemplateLanguage        = oapi.WhatsAppTemplateLanguage
	WhatsAppTemplateLanguageSummary = oapi.WhatsAppTemplateLanguageSummary
	WhatsAppTemplateLanguageList    = oapi.WhatsAppTemplateLanguageList
)

// WhatsAppTemplateComponent is one content block of a language; the other three
// carry Meta's verdict on the language that holds them.
type (
	WhatsAppTemplateComponent       = oapi.WhatsAppTemplateComponent
	WhatsAppTemplateQuality         = oapi.WhatsAppTemplateQuality
	WhatsAppTemplateRejection       = oapi.WhatsAppTemplateRejection
	WhatsAppTemplateSubmissionError = oapi.WhatsAppTemplateSubmissionError
)

// The senders a send can name: a WhatsAppNumber is one connected number and the
// state WhatsApp reports for it, WhatsAppNumberProfile is what people see about
// the business in WhatsApp, and a WhatsAppNumberEvent is one step in how the
// number reached its current state.
type (
	WhatsAppNumber          = oapi.WhatsAppNumber
	WhatsAppNumberID        = oapi.WhatsAppNumberID
	WhatsAppNumberList      = oapi.WhatsAppNumberList
	WhatsAppNumberProfile   = oapi.WhatsAppNumberProfile
	WhatsAppNumberEvent     = oapi.WhatsAppNumberEvent
	WhatsAppNumberEventList = oapi.WhatsAppNumberEventList
)

// The filters a number or event list takes. WhatsAppNumberStatus is an open
// enum, so openenums.gen.go declares it in this package already.
type (
	WhatsAppNumberScope          = oapi.WhatsAppNumberScope
	WhatsAppNumberSortField      = oapi.WhatsAppNumberSortField
	WhatsAppNumberEventSortField = oapi.WhatsAppNumberEventSortField
)

// WhatsAppBusinessAccount is the account a number reports in its own Waba
// field, carrying Meta's reviews of the business rather than of the number.
type (
	WhatsAppBusinessAccount          = oapi.WhatsAppBusinessAccount
	WhatsAppBusinessAccountList      = oapi.WhatsAppBusinessAccountList
	WhatsAppBusinessAccountSortField = oapi.WhatsAppBusinessAccountSortField
)

// PhoneNumberLookup is what we know about a phone number; EmailLookup is the
// verdict on an email address. Every block a phone lookup carries reports its
// own status, so a partial answer is visible rather than silent.
type (
	PhoneNumberLookup = oapi.PhoneNumberLookup
	EmailLookup       = oapi.EmailLookup
)

// WhatsApp statistics responses, returned by the Client.Whatsapp.Stats methods.
// Each is the read-side body for one breakdown; the Inbound set counts
// messages the workspace's own numbers received rather than what it sent.
type (
	// WhatsAppStatsSummary is the delivery and latency totals for a window,
	// optionally with a previous-period comparison. Returned by Stats.Summary.
	WhatsAppStatsSummary = oapi.WhatsAppStatsSummary
	// WhatsAppStatsResponse is a time series of per-bucket points. Returned by
	// Stats.Daily and Stats.Hourly.
	WhatsAppStatsResponse                   = oapi.WhatsAppStatsResponse
	WhatsAppStatsByErrorCodeResponse        = oapi.WhatsAppStatsByErrorCodeResponse
	WhatsAppStatsByTemplateResponse         = oapi.WhatsAppStatsByTemplateResponse
	WhatsAppStatsByTemplateCategoryResponse = oapi.WhatsAppStatsByTemplateCategoryResponse
	WhatsAppStatsByTagResponse              = oapi.WhatsAppStatsByTagResponse
	WhatsAppStatsByPhoneNumberResponse      = oapi.WhatsAppStatsByPhoneNumberResponse
	WhatsAppStatsByCountryResponse          = oapi.WhatsAppStatsByCountryResponse

	WhatsAppInboundStatsSummaryResponse       = oapi.WhatsAppInboundStatsSummaryResponse
	WhatsAppInboundStatsResponse              = oapi.WhatsAppInboundStatsResponse
	WhatsAppInboundStatsByPhoneNumberResponse = oapi.WhatsAppInboundStatsByPhoneNumberResponse
)

// WhatsAppMessageTemplateComponent is a filled-in template component — supplied
// on a template send and echoed back on the sent message.
// WhatsAppMessageTemplateComponentParameter is one of its placeholder values.
type (
	WhatsAppMessageTemplateComponent          = oapi.WhatsAppMessageTemplateComponent
	WhatsAppMessageTemplateComponentParameter = oapi.WhatsAppMessageTemplateComponentParameter
)

// The free-form content arms a send carries in place of a template. Each is the
// wire object verbatim: the SDK sugars only the template handle, because the
// server decides which single arm a send may carry and reports the verdict.
type (
	WhatsAppTextSend     = oapi.WhatsAppTextSend
	WhatsAppImageSend    = oapi.WhatsAppImageSend
	WhatsAppVideoSend    = oapi.WhatsAppVideoSend
	WhatsAppAudioSend    = oapi.WhatsAppAudioSend
	WhatsAppStickerSend  = oapi.WhatsAppStickerSend
	WhatsAppDocumentSend = oapi.WhatsAppDocumentSend
	WhatsAppLocationSend = oapi.WhatsAppLocationSend
)

// The interactive send arm and the shapes it nests. Each is the wire object
// verbatim: `Type` names the kind and the matching field carries it, and the
// server rejects a send that carries a second kind's field.
type (
	WhatsAppInteractiveSend                 = oapi.WhatsAppInteractiveSend
	WhatsAppInteractiveHeaderSend           = oapi.WhatsAppInteractiveHeaderSend
	WhatsAppInteractiveButtonSend           = oapi.WhatsAppInteractiveButtonSend
	WhatsAppInteractiveQuickReplyButtonSend = oapi.WhatsAppInteractiveQuickReplyButtonSend
	WhatsAppInteractiveCtaUrlSend           = oapi.WhatsAppInteractiveCtaUrlSend
	WhatsAppInteractiveListSend             = oapi.WhatsAppInteractiveListSend
	WhatsAppInteractiveListSectionSend      = oapi.WhatsAppInteractiveListSectionSend
	WhatsAppInteractiveListRowSend          = oapi.WhatsAppInteractiveListRowSend
	WhatsAppInteractiveCardSend             = oapi.WhatsAppInteractiveCardSend
	WhatsAppInteractiveCardHeaderSend       = oapi.WhatsAppInteractiveCardHeaderSend
)

// The contact-card send arm and the shapes it nests. Each is the wire object
// verbatim: a card's Name needs FormattedName plus at least one other part, and a
// PhoneNumber in E.164 earns the card a button that opens a chat with it.
type (
	WhatsAppContactCardSend    = oapi.WhatsAppContactCardSend
	WhatsAppContactNameSend    = oapi.WhatsAppContactNameSend
	WhatsAppContactOrgSend     = oapi.WhatsAppContactOrgSend
	WhatsAppContactPhoneSend   = oapi.WhatsAppContactPhoneSend
	WhatsAppContactEmailSend   = oapi.WhatsAppContactEmailSend
	WhatsAppContactUrlSend     = oapi.WhatsAppContactUrlSend
	WhatsAppContactAddressSend = oapi.WhatsAppContactAddressSend
)

// The kinds a send may name, and the kinds its header and buttons may name.
// Closed on the write side, unlike their read counterparts above: WhatsApp adds
// kinds we render back before we accept them on a send. The names carry the
// schemas' own `Write` infix so each constant reads against the field it sets.
type (
	WhatsAppInteractiveTypeWrite       = oapi.WhatsAppInteractiveTypeWrite
	WhatsAppInteractiveHeaderTypeWrite = oapi.WhatsAppInteractiveHeaderTypeWrite
	WhatsAppInteractiveButtonTypeWrite = oapi.WhatsAppInteractiveButtonTypeWrite
)

const (
	WhatsAppInteractiveTypeWriteButton                 = oapi.WhatsAppInteractiveTypeWriteButton
	WhatsAppInteractiveTypeWriteList                   = oapi.WhatsAppInteractiveTypeWriteList
	WhatsAppInteractiveTypeWriteCtaUrl                 = oapi.WhatsAppInteractiveTypeWriteCtaUrl
	WhatsAppInteractiveTypeWriteCarousel               = oapi.WhatsAppInteractiveTypeWriteCarousel
	WhatsAppInteractiveTypeWriteLocationRequestMessage = oapi.WhatsAppInteractiveTypeWriteLocationRequestMessage
	WhatsAppInteractiveTypeWriteRequestContactInfo     = oapi.WhatsAppInteractiveTypeWriteRequestContactInfo
)

const (
	WhatsAppInteractiveHeaderTypeWriteText     = oapi.WhatsAppInteractiveHeaderTypeWriteText
	WhatsAppInteractiveHeaderTypeWriteImage    = oapi.WhatsAppInteractiveHeaderTypeWriteImage
	WhatsAppInteractiveHeaderTypeWriteVideo    = oapi.WhatsAppInteractiveHeaderTypeWriteVideo
	WhatsAppInteractiveHeaderTypeWriteDocument = oapi.WhatsAppInteractiveHeaderTypeWriteDocument
)

const (
	WhatsAppInteractiveButtonTypeWriteQuickReply = oapi.WhatsAppInteractiveButtonTypeWriteQuickReply
	WhatsAppInteractiveButtonTypeWriteCtaUrl     = oapi.WhatsAppInteractiveButtonTypeWriteCtaUrl
)

// WhatsAppTag is a structured {name, value} label on a WhatsApp send.
type WhatsAppTag = oapi.Tag

// VoiceCall is one call-detail record, in flight or settled; VoiceCallList is a
// page of them.
type (
	VoiceCall     = oapi.VoiceCall
	VoiceCallList = oapi.VoiceCallList
)

// VoiceCallStatus is how a call ended, or that it is still ringing or connected.
// VoiceCallDirection is which way the call was placed.
type (
	VoiceCallStatus    = oapi.VoiceCallStatus
	VoiceCallDirection = oapi.VoiceCallDirection
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
	EmailStatusScheduled      EmailStatus = "scheduled"
	EmailStatusAccepted       EmailStatus = "accepted"
	EmailStatusProcessed      EmailStatus = "processed"
	EmailStatusDelivered      EmailStatus = "delivered"
	EmailStatusDeferred       EmailStatus = "deferred"
	EmailStatusBounced        EmailStatus = "bounced"
	EmailStatusComplained     EmailStatus = "complained"
	EmailStatusRejected       EmailStatus = "rejected"
	EmailStatusPartialFailure EmailStatus = "partial_failure"
	EmailStatusCanceled       EmailStatus = "canceled"
)

const (
	EmailBroadcastStatusDraft     EmailBroadcastStatus = "draft"
	EmailBroadcastStatusScheduled EmailBroadcastStatus = "scheduled"
	EmailBroadcastStatusAccepted  EmailBroadcastStatus = "accepted"
	EmailBroadcastStatusSending   EmailBroadcastStatus = "sending"
	EmailBroadcastStatusSent      EmailBroadcastStatus = "sent"
	EmailBroadcastStatusCanceling EmailBroadcastStatus = "canceling"
	EmailBroadcastStatusCanceled  EmailBroadcastStatus = "canceled"
	EmailBroadcastStatusFailed    EmailBroadcastStatus = "failed"
)

const (
	EmailBroadcastCategoryMarketing     EmailBroadcastCategory = "marketing"
	EmailBroadcastCategoryTransactional EmailBroadcastCategory = "transactional"
)

const (
	EmailSendAllowanceWindowDaily   EmailSendAllowanceWindow = "daily"
	EmailSendAllowanceWindowMonthly EmailSendAllowanceWindow = "monthly"
	EmailSendAllowanceWindowNone    EmailSendAllowanceWindow = "none"
)

const (
	EmailRecipientStatusAccepted   EmailRecipientStatus = "accepted"
	EmailRecipientStatusProcessed  EmailRecipientStatus = "processed"
	EmailRecipientStatusDeferred   EmailRecipientStatus = "deferred"
	EmailRecipientStatusDelivered  EmailRecipientStatus = "delivered"
	EmailRecipientStatusBounced    EmailRecipientStatus = "bounced"
	EmailRecipientStatusComplained EmailRecipientStatus = "complained"
	EmailRecipientStatusRejected   EmailRecipientStatus = "rejected"
)

// failure_reason is null on a broadcast that has not failed, and the field is a
// plain pointer, so nil is the test. There is no constant for the generator's
// "<nil>" rendering of the schema's null.
const (
	EmailBroadcastFailureReasonEmptyAudience       EmailBroadcastFailureReason = "empty_audience"
	EmailBroadcastFailureReasonAudienceUnavailable EmailBroadcastFailureReason = "audience_unavailable"
	EmailBroadcastFailureReasonContentInvalid      EmailBroadcastFailureReason = "content_invalid"
	EmailBroadcastFailureReasonInsufficientFunds   EmailBroadcastFailureReason = "insufficient_funds"
	EmailBroadcastFailureReasonQuotaExceeded       EmailBroadcastFailureReason = "quota_exceeded"
	EmailBroadcastFailureReasonInternalError       EmailBroadcastFailureReason = "internal_error"
)

// An absent bounce_type is how the wire expresses "unclassified", so there is
// no constant for the generator's "<nil>" rendering of the schema's null.
const (
	EmailEventBounceTypeHard         EmailEventBounceType = "hard"
	EmailEventBounceTypeSoft         EmailEventBounceType = "soft"
	EmailEventBounceTypeBlock        EmailEventBounceType = "block"
	EmailEventBounceTypeAdmin        EmailEventBounceType = "admin"
	EmailEventBounceTypeUndetermined EmailEventBounceType = "undetermined"
)

// A recipient carries its own bounce vocabulary, identical in values to the
// event's. They are separate types on the wire, so a constant from one does not
// compile against the other.
const (
	EmailRecipientBounceTypeHard         EmailRecipientBounceType = "hard"
	EmailRecipientBounceTypeSoft         EmailRecipientBounceType = "soft"
	EmailRecipientBounceTypeBlock        EmailRecipientBounceType = "block"
	EmailRecipientBounceTypeAdmin        EmailRecipientBounceType = "admin"
	EmailRecipientBounceTypeUndetermined EmailRecipientBounceType = "undetermined"
)

const (
	EmailEventRejectionReasonRecipientSuppressed EmailEventRejectionReason = "recipient_suppressed"
	EmailEventRejectionReasonTransmissionFailed  EmailEventRejectionReason = "transmission_failed"
	EmailEventRejectionReasonGenerationFailure   EmailEventRejectionReason = "generation_failure"
	EmailEventRejectionReasonPolicyRejection     EmailEventRejectionReason = "policy_rejection"
	EmailEventRejectionReasonDomainUnverified    EmailEventRejectionReason = "domain_unverified"
	EmailEventRejectionReasonQuotaExceeded       EmailEventRejectionReason = "quota_exceeded"
	EmailEventRejectionReasonRecipientNotAllowed EmailEventRejectionReason = "recipient_not_allowed"
)

const (
	EmailRecipientRejectionReasonRecipientSuppressed EmailRecipientRejectionReason = "recipient_suppressed"
	EmailRecipientRejectionReasonTransmissionFailed  EmailRecipientRejectionReason = "transmission_failed"
	EmailRecipientRejectionReasonGenerationFailure   EmailRecipientRejectionReason = "generation_failure"
	EmailRecipientRejectionReasonPolicyRejection     EmailRecipientRejectionReason = "policy_rejection"
	EmailRecipientRejectionReasonDomainUnverified    EmailRecipientRejectionReason = "domain_unverified"
	EmailRecipientRejectionReasonQuotaExceeded       EmailRecipientRejectionReason = "quota_exceeded"
	EmailRecipientRejectionReasonRecipientNotAllowed EmailRecipientRejectionReason = "recipient_not_allowed"
)

const (
	RecipientRoleTo  RecipientRole = "to"
	RecipientRoleCc  RecipientRole = "cc"
	RecipientRoleBcc RecipientRole = "bcc"
)

// Category classifies a send's suppression policy.
type Category = oapi.EmailMessageCategory

const (
	CategoryTransactional Category = "transactional"
	CategoryMarketing     Category = "marketing"
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

	// Workspace is the workspace a credential is scoped to: its id, name, and
	// organization.
	Workspace = oapi.Workspace

	WebhookAttemptList          = oapi.WebhookAttemptList
	WebhookEndpoint             = oapi.WebhookEndpoint
	WebhookEndpointCreate       = oapi.WebhookEndpointCreate
	WebhookEndpointCreated      = oapi.WebhookEndpointCreated
	WebhookEndpointID           = oapi.WebhookEndpointID
	WebhookEndpointList         = oapi.WebhookEndpointList
	WebhookEndpointUpdate       = oapi.WebhookEndpointUpdate
	WebhookRotateSecretResponse = oapi.WebhookRotateSecretResponse
	WebhookSortField            = oapi.WebhookSortField
	WebhookTestRequest          = oapi.WebhookTestRequest
	WebhookTestResponse         = oapi.WebhookTestResponse
)
