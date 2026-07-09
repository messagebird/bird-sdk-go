# Changelog

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
