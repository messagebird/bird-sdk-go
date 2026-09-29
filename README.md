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
- **`client.Lookup`** — `PhoneNumber` (what a number is: country, serving network, line type, plus paid properties named in `Type`) and `Email` (whether an address is worth sending to). Every answer is billed: the base number lookup once plus once per **delivered** property, an email lookup once per answered address. Only a property block whose `Status` is `"ok"` carries a value, and only that one is billed, so read the status before the value.
- **`client.Realtime`** — `Publish`, `PublishBatch`, plus `Channels` (`List`, `Get`, `Members`) and `Members.Disconnect`. Every call takes the Realtime app id and needs the app's own credentials on top of the API key: `option.WithRealtimeCredentials(key, secret)`, at construction or per call.
- **`client.Contacts`** — `Create`, `Get`, `Update`, `Delete`, `Batch`, `List` (auto-paginating). `client.Audiences` groups them (`Create`, `Get`, `Update`, `Delete`, `List`, plus `ListContacts`, `AddContacts`, `RemoveContacts`, `RemoveContact`), and `client.ContactProperties` defines the fields a contact carries (`Create`, `Get`, `Update`, `List`, `Archive`, `Unarchive`).
- **`client.Domains`** — `Create`, `Get`, `Update`, `Delete`, `List`, and `Verify` (check a sending domain's DNS).
- **`client.Webhooks`** — `Unwrap` (verify a signed event into a typed value).
- **Typed errors.** A failure is a `*bird.APIError` (or a richer `*bird.RateLimitError` / `*bird.ValidationError`) you branch on with `errors.As`. Transient failures (timeouts, 429, 5xx) are retried automatically with a reused idempotency key.
- **Options** configure the client and override per call (`option.WithEmailDefaults`, `WithTimeout`, `WithIdempotencyKey`, …).
- **`client.Get/Post/Put/Patch/Delete`** reach endpoints outside the curated surface.

Optional collection fields in generated request params distinguish `nil` from an empty value: `nil` omits the field, while `[]string{}` or `map[string]any{}` sends an empty collection. The API decides whether that clears a value or is rejected.

## Examples

Runnable, per-method examples live in [`example_test.go`](./example_test.go) and render under each method on [pkg.go.dev](https://pkg.go.dev/github.com/messagebird/bird-sdk-go): sending (simple and rich), error handling, get, pagination, channel defaults, the webhook receiver, and the escape hatch.

## Design

The wire types and a low-level client are generated from the OpenAPI spec into `internal/oapi`; this package is the hand-written idiomatic layer on top.

## Apple Messages quickstart

Use `client.Amb.Send` to reply to an open, customer-initiated conversation as a configured, connected business account. Set `BIRD_API_KEY` and `AMB_CONVERSATION_ID`, then run the [Apple Messages example](examples/quickstart-amb/main.go). The key needs `amb:read`, `amb:write` and `amb_management:read`.

The example sends a real reply. Verify the recipient and content before running it. `accepted` means queued; `sent` means Apple gateway acceptance, not device delivery or a read receipt. Native payment, authentication and invitation requests are not part of this public channel release.
