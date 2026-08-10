# Changelog

## 0.26.0

- An SMS message's `cost` now names its components: `transaction_amount` is what Bird charged to carry the message, and `passthrough_amount` is reserved for third-party fees such as US 10DLC carrier surcharges. `amount` remains the total.

## 0.25.0

- Listing WhatsApp messages gains a `category` filter, matching the equivalent filter on SMS and email messages.

## 0.24.1

- The Realtime app list's `sort` filter is now a named type, `RealtimeAppSortField` (`created_at` | `name`), instead of a bare string, and the Realtime app `name` field carries a description. No method or signature changes.

## 0.24.0

- Batch contact upserts now match each entry automatically on every identifier it carries (email, phone, or external_id), refusing entries whose identifiers belong to more than one contact; `match_on` (`email`, `phone`, or `external_id`) forces a single key. Result items echo what each entry supplied under a nested `entry` object (`email`, `phone`, `external_id`, null where absent, never the contact's current state), plus a top-level `matched_on` naming which identifier matched, null for created rows.
- Failed rows in a batch contact upsert carry the specific error `code` (for example `E04058`, ambiguous match, versus `E04055`, phone taken) alongside `type` and `message`, so a sync can branch on which conflict it hit.
- **Breaking (0.x):** `Contact.Channels` is removed: the field restated which identifiers are set under a reachability claim the platform cannot back. Read `Email`/`Phone` presence directly.
- **Breaking (0.x):** `Contact.Email` becomes a pointer: a contact may be identified by an E.164 phone number instead of, or as well as, an email address. Contacts gain `Phone` and the contact list gains an exact `phone` filter.
- Filter WhatsApp messages by `direction`. The unfiltered list returns the whole conversation, so `direction` narrows it to what you sent or what the contact sent you.

## 0.23.0

- Voice call webhook payloads name the two parties from and to, replacing src_number and dst_number. A handler reading those fields on voice_call.initiated, voice_call.answered or voice_call.ended must rename them; every other field is unchanged.
- WhatsApp and SMS template `language` fields now take a BCP-47 tag (for example `pt-BR`); Meta's underscore form (`pt_BR`) is still accepted as an input alias but is no longer echoed back. WhatsApp template sends can now also address a template by `id`, as an alternative to `slug`: the existing `Template` field takes either, resolving a `wat_`-prefixed value as the id — the same convention it already uses for SMS's `smt_`.
- WhatsApp send: the `to` field now accepts a business-scoped user ID as well as an E.164 phone number, so you can message a WhatsApp user whose phone number you do not have. One-time-passcode templates still require a phone number and return `422 WhatsAppRecipientNotSupportedForTemplate` when sent to a business-scoped user ID.

## 0.22.0

- Add `bird.Email(...)`, the pointer shorthand for a `format: email` field, so an email recipient such as `VerificationTo.EmailAddress` can be set inline alongside `bird.String` and `bird.Bool`.

## 0.21.1

- The five bespoke delete-blocked error codes (IPPoolContainsIPs, DomainHasMailboxes, AudienceInUse, TCRBrandHasCampaigns, ConfigurationInUse) are retired in favor of the single E01028 ResourceInUse. domains.delete may now return a 409 (Conflict), decoded as DeleteDomainResponse.JSON409, when the domain still has mailboxes bound to it.

## 0.21.0

- Add a `datetime` contact property type: an RFC 3339 timestamp with an explicit offset.
- Add realtime.members.send: deliver an event to every connection one member holds, addressing the person rather than a channel.
- Listing contacts gains an `include_total` flag for a total count, and `q` now matches first and last name as well as email.
- WhatsApp `rejected` now covers every refusal before transmit, not only a suppressed recipient: a charge decline (insufficient wallet balance, unpriced destination) and an undeliverable recipient report `status: rejected` with a `whatsapp.rejected` timeline event instead of `failed`. `whatsapp.rejected` is also now published on the message-events enum.
- **Breaking:** a WhatsApp send against a template that declares named parameters must now name every parameter, matched as a set, so order no longer matters; an unnamed, misnamed, duplicated, or missing name returns `422 WhatsAppTemplateParameterMismatch`. `bird_otp` stays positional: its values go in `{{n}}` order and must carry no name. Unnamed values were previously matched by position and could render the wrong content while reporting success, so sends made against these templates before this release are worth re-checking.
- A parameter that arrived with no description now carries the one its field documents.
- Clarify the email broadcast and template descriptions, including the errors each operation can return.
- Correct the `parameters` field description for an inline send.
- Document that an authentication-category message returns a redacted body.
- Template variables now report a `sensitive` flag showing whether the value is redacted before storage.

## 0.20.0

- BREAKING: open-enum fields are now defined types (EmailEventType, VerificationChannel, WhatsAppErrorCode and six others) instead of aliases for string, so they carry their known values and a Valid() method. Assigning a string variable to one of these fields now needs an explicit conversion; an untyped string literal and any unrecognized value both still work.

## 0.19.0

- Add a per-send `language` override for templated email: `EmailSendParams.Language` on the Go SDK, a `language` keyword argument on `client.email.send`/`send_batch` in the Python SDK, and `--language` on `bird email send`. Omitting it keeps today's behavior, sending the template's default language.
- Export the known values of every open enum (EmailEventType, VerificationChannel, WhatsAppErrorCode and the rest) as constants, so an open enum is no longer a bare string with its values buried in prose.
- Email message reads and send responses now report `requested_language` and `resolved_language`.
- Email message reads and send responses now report `template_id` and `template_version_id`. A template's live version changes each time you submit it, so the version is what identifies the wording a message was actually delivered with, and the two together fetch that content back.
- Passcode attempts can report `channel_disabled`, meaning Bird has temporarily stopped sending over that channel and the verification moved on to the next one.

## 0.18.0

- **Breaking:** a WhatsApp send addresses its template by `slug` (previously `name`), and a WhatsApp message read echoes `template.slug` (previously `template.name`). Templates carry the slug as their permanent handle; the display name is cosmetic and never affects sending.

## 0.17.0

- Email message reads now report the message as delivered. For a send that used a template, `subject` and the bodies from the message-content endpoint previously returned the template source, tokens and all, which is content no recipient received; they now return that source with the send's substitution values applied. The values themselves are exposed as a new `parameters` field so the inputs stay visible beside the result. Sends that supplied their content inline are unaffected.
- Nest mailboxes and threads under email, matching the URL and the CLI/MCP names. Client.Mailbox becomes Client.Email.Mailboxes, Client.MailboxThread becomes Client.Email.Threads, Client.MailboxThreadMessage becomes Client.Email.Threads.Messages, Client.MailboxReceiveRule becomes Client.Email.Mailboxes.ReceiveRules, and Mailbox.Compose becomes Email.Mailboxes.Messages.Create. Params and service types are renamed to match.

## 0.16.0

- Add an optional `language` to the email template send block, selecting which of the template's languages to send. The template's `on_missing_language` decides whether an unstocked language falls back or is rejected.
- Add `contacts.batch`; its `Contacts` field is typed `[]ContactCreateRequest`.
- `mailbox_thread_message.reply` now accepts the full message body, including structured `tags` (`{name, value}`) and `attachments`. In Go, `ReplyAll` is now `*bool` so an explicit `false` reaches the wire.
- `mailbox_thread.update`: `AddLabels`/`RemoveLabels` collapse to a single `Labels` field, and `ContactID` becomes `Nullable[string]` so a contact link can be cleared.
- `mailbox.update`: `ReceivePolicy` and `RetentionTier` are now the typed `*MailboxUpdateReceivePolicy` / `*MailboxUpdateRetentionTier` instead of `*string`.
- Add the `mailbox` resource: `list`, `create`, `get`, `stats`, `labels`, `restore`, `resume`, and `delete`.
- **Go:** the create body's `ReceivePolicy` and `RetentionTier` are now the typed `MailboxCreateReceivePolicy` / `MailboxCreateRetentionTier` enums instead of plain strings.
- **TypeScript:** a write whose body has no required field now defaults its params to `{}`, so `bird.mailbox.create()`, `bird.audiences.update(id)`, and `bird.domains.update(id)` are callable without a body; the unused `MailboxList` envelope export is removed.
- Add the `mailbox_thread_message` reads: `list`, `get`, `body`, and `attachments`.
- **Python:** the `client.mailbox_thread.messages` nested accessor is replaced by the top-level `client.mailbox_thread_message` (matching Go and TypeScript).
- **TypeScript:** the unused `EmailThreadMessageList` export is removed.
- Add the `mailbox_thread` reads (`list`, `get`) and `delete`.
- **Go (breaking, 0.x):** `MailboxThreadService.Delete` takes `MailboxThreadDeleteParams{Permanent bool}` instead of a positional `permanent bool`.
- **TypeScript:** `mailboxThread.delete` accepts an optional `MailboxThreadDeleteQuery`; the unused `EmailThreadList` export is removed.
- **Python:** `mailbox_thread.delete` accepts a `permanent` keyword.
- Add the `mailbox_receive_rule` resource: `list`, `create`, and `delete`.
- **Go:** the create body's `Action` is now the typed `ReceiveRuleCreateAction` enum instead of a plain string.
- **Python:** the `client.mailbox.receive_rules` nested accessor is replaced by the top-level `client.mailbox_receive_rule` (matching Go and TypeScript).
- **TypeScript:** the unused `ReceiveRuleList` export is removed.
- Add the SMS reads `get` and `list`. The Go service type is now `SmsService` (was `SMSService`), matching the client field. `SmsListParams` exposes `Status` and `ErrorCode` as `[]string` and adds the previously-missing `Tag` filter; `Category` keeps its enum type.
- Add the `SmsTemplates` and `WhatsappTemplates` resources. The list filters are now plain strings (`SMSTemplateScope` / `SMSTemplateCategory` and their constants are removed). Service types are `SmsTemplatesService` / `WhatsappTemplatesService`, matching the client fields.
- Add the WhatsApp reads `get`, `list`, and `list_events`. The Go service type is now `WhatsappService` (was `WhatsAppService`). `WhatsappListParams.Status` is now `[]string` (was `Statuses []WhatsAppMessageStatus`) and adds the previously-missing `Tag` filter.
- **Breaking (0.x):** `VerificationsService` is now `VerifyVerificationsService`; its params take the nested shape: `To VerificationTo` replaces flat `Email`/`Phone`, and `Options *VerificationOptions` replaces `CodeLength`/`Channels`.
- WhatsApp messages now return `cost`, the amount charged for the message.
- Go read filters that reference a named enum are now typed with it instead of a plain string: `SMSTemplateListParams.Scope` / `Category` are `TemplateScope` / `SMSMessageCategory`, and the email-stats breakdowns take their `EmailStatsSortMetric` / `EmailEngagementSortMetric` / `EmailMailboxProviderSortMetric` sort metric.
- The email template send block addresses a stored template by `slug` (previously `name`). Templates carry the slug as their permanent handle plus a separate free-text display `name`.
- The realtime.* webhook event type constants are no longer exported. Realtime webhooks are created and managed in the Bird dashboard.
- **Breaking:** remove the WhatsApp templates-list surface — `bird whatsapp templates list`, the `whatsapp_templates_list` MCP tool, and `whatsappTemplates.list` / `WhatsappTemplates.List` / `whatsapp_templates.list` in the TypeScript, Go, and Python SDKs. WhatsApp is still in preview and the templates contract is being reshaped for localisation; templates return to the public and command audiences at GA in the new shape.
- Add the `soft_bounce` verification attempt failure reason (open enum) for transient email bounces.
- Clarified the VoiceCallStatus documentation: ringing and in_progress describe a call that is still up, rather than values held back for a future feature.
- SMS alphanumeric sender IDs now allow dashes and underscores alongside letters, digits, and spaces, and must contain at least one letter with no separator at either end. A digits-only value is a long code or short code, and is no longer accepted as a sender ID.
- `contact-properties` `create` and `update`: the fallback value is typed as any JSON value, and the property type is the named `ContactPropertyType`.
- `contacts.update`: the `Email` field is now a plain string rather than a pointer.
- Add the email read methods: `get`, `list`, and `cancel`.
- `EmailThreadMessageReplyRequest` gains an optional `attachments` field, matching the compose and direct-send surfaces.
- The stats trend-grain, message direction, and email status read filters are now typed rather than plain strings, so each carries the values it accepts.

## 0.15.0

- Add Realtime data-plane methods: publish, batch publish, channel list/get/members, and member disconnect.
- Add the domains resource: create, read, update, delete, and verify. List supports `sort`/`order`/`include_total`; update clears the tracking domain via `bird.Null(...)` instead of a `ClearTracking` flag.
- Add `verify.*` webhook event types and payloads.
- Internal improvements.

## 0.14.0

- Audiences params now match the API's declared field types: `AudienceCreateParams.Type` is the typed enum (was `string`), and `AudienceUpdateParams.Name` is a plain `string` (the name cannot be cleared).
- `ContactUpdateParams` first name, last name, and external id can now be cleared: they are `bird.Nullable[string]` (was `*string`) — `bird.Value(v)` sets, `bird.Null[string]()` clears (explicit JSON null), zero omits. Email stays a pointer (it cannot be cleared).
- Nullable request fields can now send an explicit JSON `null` to clear a value. Clearable request params use `bird.Nullable[T]`, built with `bird.Value(v)` (set), `bird.Null[T]()` (clear), or left zero (omit); response fields stay plain pointers for ergonomic reads. `AudienceUpdateParams.Description` is now `bird.Nullable[string]` (was `*string`), and `DomainUpdateParams.ClearTracking` now sends a real null.
- The verification terminal reason is now the named type `VerificationTerminalReason`, carrying its known values, rather than a bare string.
- Stats `period.grain` is now typed as a shared `StatsGrain` (`day` | `hour`) instead of a plain string. No wire or behavioural change.

## 0.13.0

- **Breaking:** `Email.Stats.BySendingIp` is renamed `BySendingIP`, with its `EmailStatsBySendingIpParams`/`EmailStatsBySendingIpResponse` types renamed to `...IP...` — idiomatic Go initialism (golint ST1003), matching the `SendingIP` field the SDK already uses.
- Resource and package docstrings now describe behavior only.
- Internal improvements.

## 0.12.0

- Agent mailboxes (inbox.ai): create and manage durable inboxes; receive, read, label, compose, and reply over the API
- Internal improvements.

## 0.11.0

- Add the `rejected` WhatsApp message delivery status, returned when the recipient is on the workspace's suppression list.
- Add the `whatsapp.rejected` webhook event, delivered when Bird rejects a WhatsApp message before sending because the recipient is on the workspace's suppression list.
- Rename voice call webhook event types from voice.call.* to voice_call.* (single-dot resource convention; events were never emitted, so no delivered payload changes)

## 0.10.1

- Message list filters: created_after is an inclusive lower bound and created_before is now an exclusive upper bound.

## 0.10.0

- Add sms.tfn_verification webhook event types
- Email message status enum constants are now type-prefixed (Accepted -> ListEmailMessagesParamsStatusAccepted, ...) so their names are stable across releases
- Add email statistics reads under `email.stats`: the period summary, the daily and hourly time series, and the dimension breakdowns (by tag, category, sending IP, sending domain, recipient domain, mailbox provider, mailbox-provider region, template, location, client, bounce code, complaint type, and broadcast).
- **Breaking:** the Realtime webhook event type `realtime.subscription_count` is now `realtime.connection_count`, matching Bird's Realtime vocabulary (per channel it counts connections — one connection cannot subscribe twice). Realtime is in early access; the old event type had no GA consumers.
- Operations and fields now document their units, defaults, omission behavior, and per-value status meanings. Several descriptions were corrected to match actual behavior, including engagement-rate denominators, suppression prefix matching, and stored-content retention.
- Internal improvements.
- WhatsApp templates: create and list/get a workspace's own message templates. Reads now include a template id and an optional description; create takes a name, category, components, a WhatsApp language code, and an optional description; sending gained a named parameter name for named-parameter templates. Additive; no breaking change.

## 0.9.4

- `make test` no longer enables the race detector by default; the monorepo's CI still runs the suite with `-race`. Run `go test -race ./...` to opt in. No library code changed.

## 0.9.3

- Suppressions: `reason`, `origin`, and `applies_to` are now documented as growing vocabularies (open enums on the wire) — `origin` gained `unsubscribe_link`, a suppression created by the recipient through Bird's hosted unsubscribe page or its one-click link. Treat unknown values as informational rather than rejecting the record. Additive; no breaking change.

## 0.9.2

- Add voice call-event webhook types: `voice.call.initiated`, `voice.call.answered`, and `voice.call.ended` are now recognized event types with typed payloads. Additive; no breaking change.

## 0.9.1

- Documentation search results now carry a `Slug`, and a new `DocsPage` type describes a documentation page's full Markdown. Additive wire types for the public docs read/search operations; no new SDK method.

## 0.9.0

- **Breaking:** WhatsApp message reads now return `from` and `to` (each a phone number and/or business-scoped user ID) in place of `business` and `contact`, matching the SMS/email convention.

## 0.8.4

- **Breaking:** the contact list free-text filter is now `ContactListParams.Q` (was `Search`), matching the API's renamed `q` query parameter. Rename `Search:` to `Q:` at call sites.

## 0.8.3

- Documentation clarifications.

## 0.8.2

- Received messages and the `email.received` event now carry `Authentication` (`pass`/`fail`/`unknown`), a single summary of sender authentication; treat `unknown` as not verified. The `SpfPass`/`DkimPass`/`DmarcPass` fields remain. Additive; no breaking change.

## 0.8.1

- Add the WhatsApp webhook event types: `EventTypeWhatsappAccepted`, `EventTypeWhatsappSent`, `EventTypeWhatsappDelivered`, `EventTypeWhatsappRead`, and `EventTypeWhatsappFailed`. Additive; no breaking change.

## 0.8.0

- Add the sending domains collection: `Domains.Create`, `.Get`, `.List`, `.Update`, `.Delete`, and `.Verify`. Register a sending domain, publish the DNS records it returns, then verify until it is usable as a sender. Requires an API key with the `domains` scope.

## 0.7.6

- Clarify that `DocsSearchResult.Url` and `.DocUrl` are absolute URLs, matching `.MarkdownUrl` and the API's actual output. Documentation only; no API or behavior change.

## 0.7.5

- Add the Realtime webhook event types: `EventTypeRealtimeCacheChannels`, `EventTypeRealtimeChannelExistence`, `EventTypeRealtimeClientEvents`, `EventTypeRealtimePresence`, and `EventTypeRealtimeSubscriptionCount`. Additive; no breaking change.

## 0.7.4

- Contacts now carry `channels` (the channels a contact can be reached on) and audience members carry the `audiences` they belong to. Listing an audience's contacts gains an optional `search` filter (email substring). Additive response fields and an optional parameter; no breaking change.

## 0.7.3

- Correct the `Verify.Verifications.Check` documentation: an already-resolved verification is no longer checkable and returns a 404, not a result with `Success` false. Documentation only; no API or behavior change.

## 0.7.2

- WhatsApp failure detail now carries `MetaErrorCode`, the raw error code from the WhatsApp Cloud API, and a fuller `Description` sourced from Meta's error details. Additive response fields; no breaking change.

## 0.7.1

- Correct the error-code names shown in preview-feature field descriptions (regenerated from the API spec). Documentation only; no API or behavior change.

## 0.7.0

- Add the Verify product: `Verify.Verifications.Create` sends a one-time passcode to a recipient and `Verify.Verifications.Check` validates the code they submit.

## 0.6.0

- Add the WhatsApp channel: `Whatsapp.Send`, `.Get`, `.List`, `.ListEvents`. Add WhatsApp templates (read-only): `WhatsappTemplates.List`.

## 0.5.0

- Remove the email templates collection (`EmailTemplates.Create`, `.Get`, `.Update`, `.Delete`, `.Publish`, `.List`, `.ListVersions`, `.GetVersion`), added in 0.3.0. Template management is no longer part of the public API. Sending a published template with `Email.Send` (set `Template` to an `emt_…` ID or name handle) is unchanged.

## 0.4.1

- Add `Email.Cancel`: cancel a scheduled message before it sends. A message that already started sending, or was already canceled, returns a conflict error.
- Attribute the calling tool on every request via the `Bird-Caller` header, detected from the environment (no configuration).

## 0.4.0

- Add the contacts collection: `Contacts.Create`, `.Get`, `.List`, `.Update`, `.Delete`, and `.Batch` (bulk upsert by email). Requires an API key with the `email_marketing` scope.
- Add the audiences collection: `Audiences.Create`, `.Get`, `.List`, `.Update`, `.Delete`, plus membership `.ListContacts`, `.AddContacts`, `.RemoveContacts`, `.RemoveContact`.
- Add contact properties: `ContactProperties.Create`, `.Get`, `.List`, `.Update`, `.Archive`, `.Unarchive`.

## 0.3.0

- Add the SMS channel: `Sms.Send`, `Sms.SendBatch`, `Sms.Get`, `Sms.List`.
- Add SMS templates (read-only): `SmsTemplates.List`, `SmsTemplates.Get`.
- Add email templates: `EmailTemplates.Create`, `.Get`, `.Update`, `.Delete`, `.Publish`, `.List`, plus versions `.ListVersions` and `.GetVersion`.
- `Email.Send` can send a published template: set `Template` (an `emt_…` ID or name handle) with `Parameters` in place of inline `Subject`/`HTML`/`Text`.

## 0.2.2

- Rename the anonymous client-identity headers from `X-Bird-*` to `Bird-*` (the `X-` prefix is deprecated, RFC 6648). Same telemetry, new header names; no other behavior or API-surface change.

## 0.2.1

- Send anonymous `X-Bird-*` client-identity headers (surface, version, language, os, arch) on every request, so Bird can attribute API usage by surface. No personal data, credentials, or request content: just which Bird client and platform. Telemetry only; no behavior or API-surface change.

## 0.2.0

- Add batch email send: `Email.SendBatch`.

## 0.1.0

- Initial release: email send, webhook verification, pagination, typed errors.
