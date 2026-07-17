# Changelog

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
