# Changelog

## 0.51.0

- The `whatsapp.accepted`, `whatsapp.sent`, `whatsapp.delivered`, `whatsapp.read`, `whatsapp.failed`, and `whatsapp.rejected` webhook payloads now type `in_reply_to_message_id`, naming the message an outbound send answers.

## 0.50.0

- The `whatsapp.received` webhook payload now types `in_reply_to_message_id` and `interactive_reply`, naming the message an inbound reply answers and what the contact tapped.
- `rejection_reason` on a voice call record gains `number_ownership_not_verified`, returned when the country that issued a number you bought has not accepted the documents proving your workspace owns it.

## 0.49.0

- A voice call record now carries `route`: what your number did with an incoming call, whether it went to a SIP trunk, was forwarded, or was refused. Absent on outbound calls, and on calls recorded before this release, so treat a missing `route` as unknown rather than as proof the call was outbound.
- A WhatsApp message's `from` now carries the contact's `username` and `display_name` when WhatsApp discloses them. Both are descriptive; a message cannot be addressed by either. What a message is addressed by is unchanged: the contact's `phone_number`, or the `bsuid` that a contact who has not shared a phone number already carried.
- `whatsapp.send` gains `interactive`, which takes reply buttons, a list menu, a link button, media cards, or a request for the recipient's location or contact details, and `in_reply_to_message_id`, which quotes an earlier message from the same conversation. A message read back carries `interactive` and, when the contact tapped something, `interactive_reply`.

## 0.48.0

- Mailboxes now report `size_bytes`: the stored bytes of their retained messages (the metadata and extracted text kept for the retention tier plus attachment bytes). Plans can cap it per mailbox; a send from a mailbox at its cap is rejected with `E17049`.

## 0.47.0

- Mailbox `retention_tier` now accepts `90d` and `1y` on create and update, gated by the organization's plan; a tier the plan does not include is rejected with `E17048`.

## 0.46.0

- **Breaking:** the webhook verifier type is renamed to `WebhooksService`; update code that names the `WebhookService` type to the new name.
- The client is constructible with only the webhook secret — no API key — for receiver-only deployments; an API call on such a client fails with a missing-API-key error before any request is sent.
- `webhooks` now manages endpoints: list, get, create, update, delete, send a test event, inspect delivery attempts, and rotate the signing secret.

## 0.45.1

- SMS segment docs: the UCS2 limit is 70 UTF-16 code units, not 70 characters, so 35 non-BMP emoji fill one segment.

## 0.45.0

- The numbers list's `prefix` filter now matches without `country_code`: the digits are compared from the start of each stored number, ignoring a leading `+`, so short codes and a full number pasted as digits both resolve. With `country_code` it anchors behind the dial code as before.

## 0.44.0

- Add the `workspace` resource: `get()` returns the workspace the credential is scoped to, including its id, name, owning organization's id, and notification and logo settings.

## 0.43.0

- A `404` from checking or advancing a verification now says that no verification matched the recipient: either none exists for the exact recipient given, or the most recent one is already resolved as verified, expired, or out of attempts. It also says when to create another — create one when the recipient has no verification, or the last one expired or ran out of attempts, and not when the recipient already passed. The error code is unchanged, so match it exactly as before; a `404` is not evidence the recipient failed to verify, so treat your own record of an earlier `success: true` as the outcome.

## 0.42.0

- A WhatsApp message carrying contact cards a contact shared now reads them back on `contact_cards`, where it previously reported `unsupported` with type `contacts`.
- `undelivered` now documents as a final status, `validity_period` as bounding the carrier's delivery attempts rather than ours, and `scheduled` and `canceled` as reserved for send-later scheduling. `category` no longer describes per-country compliance rules that are not implemented.

## 0.41.1

- Sending a WhatsApp template: the `parameters` field now documents which button types take no value, so no component is sent for them.

## 0.41.0

- Add the `preferences` resource — list, get, create, and delete stated messaging preferences (consent grants and opt-outs) across email, SMS, and WhatsApp — plus `contacts.preferences.list` for the rows keyed to a contact's own handles, and typed webhook events for `preference.granted`, `preference.revoked`, `preference.deleted`, and `whatsapp_suppression.created`.

## 0.40.3

- `refresh_cursor` fetches the items that sort before a response's first row in the current sort order. Those are the items added since the response only when new items sort first, so a list sorted another way is refreshed by re-fetching it.

## 0.40.2

- `Number.kind` documentation now distinguishes workspace-owned subscriptions from Bird-managed shared numbers.

## 0.40.1

- Backward paging now documents `refresh_cursor` alongside `prev_cursor` as an accepted anchor, so the cursor that re-reads a list from its leading edge is discoverable without cross-referencing the response schema.

## 0.40.0

- Voice call records carry a `tags` array, and the voice call list takes a repeatable `tag` filter (`name` or `name:value`). On the CLI that is `bird voice list --tag campaign:spring --tag queue`.

## 0.39.0

- Add the `info` SMS keyword operation, and the previously undocumented `confirm`, to the SMSKeywordOperation open enum.

## 0.38.0

- Add the `sms_suppression.created` webhook event type.

## 0.37.0

- **Breaking:** the API now rejects a request that carries a query parameter an operation does not declare, with a 422 `E01029`, instead of silently ignoring it. This affects `extra_query`/`default_query` on `bird.Client`/`bird.AsyncClient` (Python), the `query` option on `bird.request` (Node), the `$query` argument on generated resource methods (PHP), and a query string written into the path passed to `Client.Get` and its sibling verb methods (Go): an unrecognized key that previously had no effect now fails the call. Check any code passing one of these against the operation's documented query parameters.
- Listing WhatsApp messages now takes `to` and `from` filters (`from_` in Python, where `from` is reserved), each matching an E.164 phone number or a business-scoped user ID. The `phone_number` filter is deprecated; use the new pair instead.

## 0.36.2

- `EmailSendParams.Category` now documents the real default: a send that leaves it unset is `marketing`, which every suppression reason holds back, so operational mail (password resets, receipts) has to set `transactional` to deliver through complaint and unsubscribe suppressions. Sends written against the previously documented `transactional` default are worth re-checking.

## 0.36.1

- A voice call's `actor` now documents the values its `type` can take: `user`, `oauth_token`, `api_key`, `system`, `sso`, and `service_account`. The enum stays open, so treat an unrecognized value as a newer type rather than an error.

## 0.36.0

- Add the `verify.verification.failed` webhook event, which fires when no planned channel could deliver a verification's passcode, plus the `undeliverable` session reason and the `not_billable` attempt reason that say why it could not.

## 0.35.0

- Email broadcasts report `started_at` and `canceled_at`: when sending began, and when cancellation was requested.

## 0.34.0

- Email click events now carry `link_name`, the name of the clicked link when the message named it.
- Numbers are now available on the public API. Search a country's available numbers, order one, list and read the numbers your workspace holds, and release a dedicated number you no longer want.
- **Breaking:** a voice call's `cost` is its own call-cost type rather than the shared money type, so a caller that names the money type for that field changes it to the call-cost type — reading `cost.amount` and `cost.currency_code` needs no change. The money type is gone from the package, having been reachable only through that field.
- A voice call's `cost` gains `outbound_amount`, `inbound_amount` and `call_handling_amount`, naming the components behind the total so a call charged for more than one thing can report which amount is which. Each reads `null` until that component is priced. Every amount on a voice call's cost now renders at six decimal places, the scale the reference documents.
- A WhatsApp message that failed because WhatsApp could not fetch the media URL, or refused the file it found there, now reports `media_rejected` on `last_error.code` instead of `undeliverable`, with WhatsApp's own reason in `last_error.description`.
- Email events now report the recipient's mailbox provider and provider region (for example `gmail`, `NA`), when the receiving mail system could be classified.
- A sending domain's inbound MX records now report `optional` as `true` until you enable receiving on that domain, so the records you must publish to send are no longer listed alongside ones that would reroute the domain's existing mail.
- The numbers reference and examples describe a number as allocated to a workspace, the same word the API reference and the dashboard use.
- The numbers methods document what each one does to a carrier and to your balance — that a search can go stale before an order lands, that an order may settle after the call returns, and that a release is irreversible.
- A paid property's `status` now documents `inconclusive` as what it is: an answer arrived that does not resolve the property, either because the number is outside the coverage of the data behind it or because the source returned a value the property does not report. It previously read as though such a value was awaiting a fix rather than being a real answer. Only `ok` ever carries a value, and only `ok` is billed.
- The WhatsApp send examples now fill the template body placeholder that bird_otp declares, so the example they show is one the API accepts rather than a 422 WhatsAppTemplateParameterMismatch.
- A voice call's `actor.type` can be `service_account`, meaning the call was placed by an integration acting for the workspace rather than by a user or an API key. Documented on the actor's `id` and `type`, and on the call's own `actor` field, which previously described only users and API keys.
- A voice call's `status` now says what `rejected` and `failed` each mean: `rejected` is a call refused rather than attempted, and `failed` is one that was attempted and did not work, with `sip_response_code` carrying what came back. `busy` and `canceled` are reserved for incoming calls and are not emitted yet, so both arrive as `failed`.
- Sending a WhatsApp message from a number whose setup has not finished now returns a `422` `WhatsAppSenderNotConnected`, rather than an unknown-number error or an accepted send that later fails.

## 0.33.0

- Add bird.Sms.Stats and bird.SmsSuppressions: SMS statistics (outbound, inbound breakdowns) and the suppression list, plus Sms.ListEvents for a message's timeline.
- Add bird.SmsKeywordRules: read Bird's SMS keyword catalogue and manage a workspace's own overrides of it.
- Add end-to-end encrypted channels: `Realtime.Publish` and `Realtime.PublishBatch` seal `private-encrypted-` payloads under `option.WithRealtimeEncryptionMasterKey`, and the new `Realtime.AuthorizeChannel` signs channel subscriptions, adding the channel's `shared_secret` on encrypted channels.
- A sending domain read now says what to do next. `next` carries the steps that get the domain verified, in the order to take them: an `external` step for a DNS record only you can publish, then the operation that re-checks it. An empty list means nothing is asked of you on that axis, and `capabilities` still reports what each capability needs before the domain can send or receive.
- Email defaults now cover `IpPoolID`, applied to every send that does not name a pool.
- `Email.Mailboxes.Messages.Create` now honors the client's email defaults for the fields a compose accepts: `ReplyTo`, `Category`, `Tags`, and `Metadata`.
- `EmailMailboxesMessagesCreateParams` gains `Tags`.
- Optional fields on a WhatsApp message's content — a media caption, a location's name and address — are omitted when unset rather than returned as null.
- Getting or listing a WhatsApp message now returns the free-form `text`, `image`, `video`, `audio`, `sticker`, `document` or `location` content it carried.
- The `whatsapp.received` webhook event is now available, delivering an inbound WhatsApp message's content as it arrives.
- `Whatsapp.Send` now accepts free-form `Text`, `Image`, `Video`, `Audio`, `Sticker`, `Document` and `Location` content, `From` for the business number every send but a Bird-managed template requires, and per-message `Tags` and `Metadata`.
- Method, parameter and field documentation now states what each call does and what each value accepts, including its default, its units, and the fields it depends on.

## 0.32.0

- Add the `EmailStatusScheduled` and `EmailStatusCanceled` constants, so a scheduled or canceled email status compares against a named value rather than a string literal.
- **Breaking:** an SMS template is referenced by its `slug`, not its `name`. On a send, `template.name` becomes `template.slug` (unchanged if you pass an id). On a template read, `slug` is the handle you send by and `name` is now the display label the old `description` carried, so `description` reads null for Bird's built-in templates.
- An SMS template read now reports what a send will do with languages, the way email and WhatsApp already did: `default_language`, `languages`, `on_missing_language` and `language_source_required`. Its `published_version_id` is now `live_version_id`, and its status vocabulary drops `approved`, which was never returned, for `inactive`.
- **Breaking:** the contacts list `phone_number` filter now takes a list of numbers rather than one, matching any of up to 50 in a single request; pass an existing single number as a one-element list, and set `limit` to at least the number of values you pass.
- **Breaking:** the recovery steps on an error are `NextAction` rather than `ErrorNextAction`, and each one states a `kind` of `operation`, `external`, `wait` or `terminal`. Rename the type at your call site, and read `kind` before `operation`: only an `operation` step carries one. Treat a `kind` you do not recognise as display-only.
- Listing WhatsApp message events now rejects an empty type filter instead of ignoring it, matching the email events filter; an unset `Type` is still omitted from the request, so it keeps returning the full timeline.
- Schedule an email from the SDK instead of dropping to raw HTTP for that one call: `EmailSendParams.ScheduledAt` holds a send until a future instant.
- Telegram is a known verification channel, so a verification can be created with `telegram` in `options.channels` and a passcode can be delivered over it.
- **Breaking:** the `type` on a WhatsApp message event, and the `type` filter when listing them, are now a defined `WhatsAppEventType` carrying its known values rather than a plain string; assigning a string variable to either now needs an explicit conversion.
- SMS sends normalize the recipient to canonical E.164 and echo that form back; an unroutable recipient returns the new SMSInvalidRecipient error.
- Voice call filters now read a number given without a leading `+` as an international one, so `from` and `to` select the same calls whichever way the number is spelled.

## 0.31.0

- Adds the lookup channel: client.Lookup.Email and client.Lookup.PhoneNumber.
- The contacts documentation now names `phone_number`, the identifier the API accepts, where it previously named `phone` — including the batch `match_on` value, which is rejected as `phone`.

## 0.30.0

- **Breaking:** a contact's phone identifier is now named `phone_number`, matching the rest of the API. It replaces `phone` in create, update, batch-upsert and read bodies; the `GET /v1/contacts` filter becomes `?phone_number=`; the `identifier` filter value and the batch `match_on` / `matched_on` values become `phone_number` — they name the field, so they move with it. On the TCR brand surface, a brand's own `phone` becomes `phone_number` as well. The CLI's `bird contacts create|update` take `--phone-number`, and `bird contacts list` filters on `--phone-number`. Qualified compounds are unchanged: `mobile_phone`, `primary_phone` and `business_contact_phone` keep their names.
- Add `options.smart_encoding` to an SMS send (`--smart-encoding` on the CLI): it replaces characters outside the GSM-7 alphabet (curly quotes, dashes, ellipses, fullwidth forms, non-breaking spaces) with their closest equivalent, which lowers the segment count and the cost on a body that would otherwise send as `UCS2`. Off unless you ask for it, and all-or-nothing: a body still holding an emoji or a non-Latin script afterwards is sent exactly as you supplied it. The message reports the settings that were applied, and its `text` reflects the body as sent.
- **Breaking:** a verification's email recipient is now named `email`, replacing `email_address` — rename it at the call site on a verification's `to`, and in any handler reading the `to` of a `verify.*` webhook payload. Phone recipients are unchanged: `phone_number` keeps its name. The old spelling is not accepted: a request sending `email_address` is rejected as an unknown property.

## 0.29.0

- **Breaking:** `carrier` and `mcc_mnc` are omitted from `sms.sent`, `sms.delivered` and `sms.received` webhook payloads when the carrier reports none, instead of arriving as `null`; `subject` on `sms.received` behaves the same way. They are now optional rather than required, so a typed field changes to a pointer or an optional — check for absence where you checked for `null`. This matches how the message resource has always reported the same fields.
- `sms.accepted` now carries `segments`, the segment breakdown the send is billed on, so a webhook-only integration can explain the `cost` on the same event instead of fetching the message to reconcile a charge.
- Subscribe to `sms.received` to be pushed inbound SMS instead of polling for it. The payload carries the message body, its segment breakdown, both numbers, and the sending operator where the carrier reports one; a `STOP` arrives as an ordinary received message and is yours to act on.

## 0.28.1

- A verification attempt can now report `delivery_timeout` as its failure reason, meaning no delivery confirmation arrived before the channel's timeout and the verification failed over to the next channel.

## 0.28.0

- The SMS error reasons known at this version are now exported as `SMSErrorCode*` constants, and the value is an open enum: a reason added by a newer server deserializes unchanged, so compare against them with a `default` branch rather than treating the set as closed.
- The `sms.expired` webhook payload now carries `error`, the same failure detail the other terminal SMS events carry: a Bird-stable `code`, a `description`, and the provider's `carrier_error_code` when it sent one.
- **Breaking:** An email template reads the recipient's contact record through `bird.contact.<attribute>` and the unsubscribe link through `bird.unsubscribe_url`. Rewrite a template that uses any other spelling for those values and republish it.
- **Breaking:** A send by template must supply a value for every parameter its template uses, and a parameter name must be a single word. A send that omits one is rejected rather than delivered with a blank in place of the value.
- **Breaking:** A `marketing` template has to place `bird.unsubscribe_url` in its body to publish, in every language it carries. The token stands on its own: no filters, not inside an `{% if %}` or `{% for %}` block, and not in the subject. Where a language has both an HTML and a text body, the HTML one has to carry it.
- `bird` is now the only name you cannot use for a parameter, so `contact`, `unsubscribe_url` and `first_name` are all available.
- Listing calls now accepts in-flight and final statuses together in one `status` filter and returns them as a single page, where mixing them used to be rejected.
- **Breaking:** a WhatsApp template parameter's `text` is now optional and should be read as nullable. It carries a value only on a `text` parameter, and a parameter of any other kind carries its value in the field named for that kind.
- Template parameters can now describe more than text. `type` accepts `image`, `video`, `gif`, `document` and `location` alongside `text`: `image`, `video`, `gif` and `document` carry a media header's file in `url`, and `location` carries a location header's point in `location`. A parameter's kind names only the field its value travels in — a coupon button's code is a plain string, so it is still a `text` parameter (`{"type":"text","text":"LUCAS25"}`), the same shape the code was authored under. A `carousel` component carries its values per card, in `cards[]`, one entry per card in the order the template was approved with.
- Three WhatsApp field descriptions now match what the API does. Omitting a template send's `language` sends the template's default language rather than returning a `422`; `received` is documented as an inbound message's status rather than as reserved; and a WhatsApp message's `cost` explains that an inbound message is never priced.

## 0.27.0

- Listing contacts gains an `identifier` filter (`email` or `phone`), and each contact now includes its `audiences`.
- **Breaking:** an SMS message's `text` is now optional — a message you sent always carries one, but a received message may not. Handle its absence rather than assuming every message has a body.
- Every `sms.*` webhook payload now carries `cost`, split into `transaction_amount` and `passthrough_amount` over one currency. The figure is as of that event, and the components are named so a subscriber merges them per component rather than replacing the object, since webhook delivery is not ordered.
- Verify gains a next-channel action: when a recipient reports the passcode never arrived, send a fresh one on the next channel in the verification's plan without waiting out the resend cooldown. Identify the verification by the same recipient you started it with, as with a check — there is still no id to store. Every passcode already sent stays valid, so a late arrival can still be checked. Available as `verify.verifications.nextChannel` / `NextChannel` / `next_channel` / `nextChannel` on the TypeScript, Go, Python, and PHP SDKs, `bird verify verifications next-channel` on the CLI, and the `verify_verifications_next_channel` MCP tool. A verification whose channel plan is exhausted answers `422 NoNextChannel`.
- Add a `voice` resource for reading the call log: list the workspace's calls with the dashboard's filters, and fetch one call at any point in its lifecycle.
- A voice call now reports `actor`, the API key or user that placed it. It is absent on calls that ended before Bird began recording it, and on any call your trunk admitted by source IP address, since that path carries no credential to identify a caller.
- A WhatsApp message's `cost` now names its components, matching SMS: `transaction_amount` is what Bird charged to send the message, and `passthrough_amount` is reserved for third-party fees. `amount` remains the total.
- The create-verification docs no longer name SMS as the phone channel: a phone recipient is verified over the phone channels enabled for its destination country, in that country's configured order.

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
