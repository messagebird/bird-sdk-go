# Changelog

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
- Type the stats trend-grain, message direction, and email status read filters, which were plain strings, by factoring each into a shared schema.

## 0.15.0

- Add Realtime data-plane methods: publish, batch publish, channel list/get/members, and member disconnect.
- Add the domains resource: create, read, update, delete, and verify. List supports `sort`/`order`/`include_total`; update clears the tracking domain via `bird.Null(...)` instead of a `ClearTracking` flag.
- Add `verify.*` webhook event types and payloads.
- Internal improvements.

## 0.14.0

- Audiences params now match the API's declared field types: `AudienceCreateParams.Type` is the typed enum (was `string`), and `AudienceUpdateParams.Name` is a plain `string` (the name cannot be cleared).
- `ContactUpdateParams` first name, last name, and external id can now be cleared: they are `bird.Nullable[string]` (was `*string`) — `bird.Value(v)` sets, `bird.Null[string]()` clears (explicit JSON null), zero omits. Email stays a pointer (it cannot be cleared).
- Nullable request fields can now send an explicit JSON `null` to clear a value. Clearable request params use `bird.Nullable[T]`, built with `bird.Value(v)` (set), `bird.Null[T]()` (clear), or left zero (omit); response fields stay plain pointers for ergonomic reads. `AudienceUpdateParams.Description` is now `bird.Nullable[string]` (was `*string`), and `DomainUpdateParams.ClearTracking` now sends a real null.
- Extract the verification terminal-reason enum into a shared `VerificationTerminalReason` type (no behavior change).
- Stats `period.grain` is now typed as a shared `StatsGrain` (`day` | `hour`) instead of a plain string. No wire or behavioural change.

## 0.13.0

- **Breaking:** `Email.Stats.BySendingIp` is renamed `BySendingIP`, with its `EmailStatsBySendingIpParams`/`EmailStatsBySendingIpResponse` types renamed to `...IP...` — idiomatic Go initialism (golint ST1003), matching the `SendingIP` field the SDK already uses.
- Docs: resource and package docstrings now describe behavior only, without internal implementation notes.
- Internal improvements.

## 0.12.0

- Agent mailboxes (inbox.ai): create and manage durable inboxes; receive, read, label, compose, and reply over the API
- Internal pipeline improvements.

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
- Documentation-only: docstrings and help text regenerated from a description pass across the entire API spec. Operations and fields now document units, defaults, omission behavior, and per-value status meanings. Several descriptions were corrected to match actual behavior, including engagement-rate denominators, suppression prefix matching, and stored-content retention. No functional changes.
- Regenerate from the beak codegen toolchain (generator provenance headers only; no API changes)
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
