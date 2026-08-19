# Bird Go SDK

The official Go SDK for the [Bird](https://bird.com) API: email, SMS, WhatsApp, verification, and Realtime, over one typed client.

```bash
go get github.com/messagebird/bird-sdk-go
```

Requires Go 1.24+.

> This SDK is generated from Bird's public OpenAPI bundle inside Bird's internal monorepo, which is the single source of truth; this repository tracks tagged releases. Generation runs in the monorepo, so `make generate` won't work from a clone here — see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Overview

`bird.NewClient(option.WithAPIKey(...))` returns a client whose region is inferred from the API key's prefix (`bk_{region}_…`); pass `option.WithBaseURL` or `option.WithRegion` to override. From there:

- **`client.Email`** — `Send`, `Get`, `List` (auto-paginating; `ListPage` for manual cursors).
- **`client.Sms`** — `Send` (free text or a stored template), `SendBatch`, `Get`, `List` (auto-paginating; `ListPage` for manual cursors). `client.SmsTemplates` (`List`, `Get`) browses the templates a send can name.
- **`client.Whatsapp`** — `Send` (a template, or free-form text/media/location), `Get`, `List` (auto-paginating; `ListPage` for manual cursors), `ListEvents` (a message's delivery timeline). Browse your workspace's approved templates in the Bird dashboard.
- **`client.Verify`** — `Verifications.Create` (send a one-time passcode) and `Verifications.Check` (validate the code a recipient submitted).
- **`client.Realtime`** — `Publish`, `PublishBatch`, plus `Channels` (`List`, `Get`, `Members`) and `Members.Disconnect`. Every call takes the Realtime app id and needs the app's own credentials on top of the API key: `option.WithRealtimeCredentials(key, secret)`, at construction or per call.
- **`client.Contacts`** — `Create`, `Get`, `Update`, `Delete`, `Batch`, `List` (auto-paginating). `client.Audiences` groups them (`Create`, `Get`, `Update`, `Delete`, `List`, plus `ListContacts`, `AddContacts`, `RemoveContacts`, `RemoveContact`), and `client.ContactProperties` defines the fields a contact carries (`Create`, `Get`, `Update`, `List`, `Archive`, `Unarchive`).
- **`client.Domains`** — `Create`, `Get`, `Update`, `Delete`, `List`, and `Verify` (check a sending domain's DNS).
- **`client.Webhooks`** — `Unwrap` (verify a signed event into a typed value).
- **Typed errors.** A failure is a `*bird.APIError` (or a richer `*bird.RateLimitError` / `*bird.ValidationError`) you branch on with `errors.As`. Transient failures (timeouts, 429, 5xx) are retried automatically with a reused idempotency key.
- **Options** configure the client and override per call (`option.WithEmailDefaults`, `WithTimeout`, `WithIdempotencyKey`, …).
- **`client.Get/Post/Put/Patch/Delete`** reach endpoints outside the curated surface.

## Examples

Runnable, per-method examples live in [`example_test.go`](./example_test.go) and render under each method on [pkg.go.dev](https://pkg.go.dev/github.com/messagebird/bird-sdk-go): sending (simple and rich), error handling, get, pagination, channel defaults, the webhook receiver, and the escape hatch.

## Design

The wire types and a low-level client are generated from the OpenAPI spec into `internal/oapi`; this package is the hand-written idiomatic layer on top.
