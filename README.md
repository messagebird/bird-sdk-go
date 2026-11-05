# Bird Go SDK

The official Go SDK for the [Bird](https://bird.com) API: email, SMS, WhatsApp, verification, and Realtime, over one typed client.

```bash
go get github.com/messagebird/bird-sdk-go
```

Requires Go 1.24+.

> This SDK is generated from Bird's public OpenAPI bundle inside Bird's internal monorepo, which is the single source of truth; this repository tracks tagged releases. Generation runs in the monorepo, so `make generate` won't work from a clone here — see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Overview

`bird.NewClient(option.WithAPIKey(...))` returns a client whose region is inferred from the API key's prefix (`bk_{region}_…` or `bm_{region}_…`); pass `option.WithBaseURL` or `option.WithRegion` to override. `option.WithClientCertificate(certFile, keyFile)` presents a client certificate on every request and resolves the region to its mTLS endpoint; it cannot be combined with `option.WithHTTPClient`. From there:

- **`client.Email`** — `Send`, `Get`, `List` (auto-paginating; `ListPage` for manual cursors).
- **`client.Sms`** — `Send` (free text or a stored template), `SendBatch`, `Get`, `List` (auto-paginating; `ListPage` for manual cursors). `client.SmsTemplates` (`List`, `Get`) browses the templates a send can name.
- **`client.Whatsapp`** — `Send` (a template, or free-form text/media/location), `Get`, `List` (auto-paginating; `ListPage` for manual cursors), `ListEvents` (a message's delivery timeline). Browse your workspace's approved templates in the Bird dashboard.
- **`client.Verify`** — `Verifications.Create` (send a one-time passcode) and `Verifications.Check` (validate the code a recipient submitted).
- **`client.Lookup`** — `PhoneNumber` (what a number is: country, serving network, line type, plus paid properties named in `Type`) and `Email` (whether an address is worth sending to). Every answer is billed: the base number lookup once plus once per **delivered** property, an email lookup once per answered address. Only a property block whose `Status` is `"ok"` carries a value, and only that one is billed, so read the status before the value.
- **`client.Realtime`** — `Publish`, `PublishBatch`, plus `Channels` (`List`, `Get`, `Members`) and `Members.Disconnect`. Every call takes the Realtime app id and needs the app's own credentials on top of the API key: `option.WithRealtimeCredentials(key, secret)`, at construction or per call.
- **`client.Contacts`** — `Create`, `Get`, `Update`, `Delete`, `Batch`, `List` (auto-paginating). `client.Audiences` groups them (`Create`, `Get`, `Update`, `Delete`, `List`, plus `ListContacts`, `AddContacts`, `RemoveContacts`, `RemoveContact`), and `client.ContactProperties` defines the fields a contact carries (`Create`, `Get`, `Update`, `List`, `Archive`, `Unarchive`).
- **`client.Domains`** — `Create`, `Get`, `Update`, `Delete`, `List`, and `Verify` (check a sending domain's DNS).
- **`client.Webhooks`** — `Unwrap` (verify a signed event into a typed value).
- **Typed errors.** A failure is a `*bird.APIError` (or a richer `*bird.RateLimitError` / `*bird.ValidationError`) you branch on with `errors.As`. An error from the Bird API carries a readable `Message`, a stable `Code` such as `E01001`, and a `DocURL` to that code's page, which says what went wrong and what to do. Every code is listed at [bird.com/docs/api/errors](https://bird.com/docs/api/errors). When a response has no code or docs link, such as one from a proxy, `Code` and `DocURL` are empty. Transient failures (timeouts, 429, 5xx) are retried automatically with a reused idempotency key.
- **Options** configure the client and override per call (`option.WithEmailDefaults`, `WithTimeout`, `WithIdempotencyKey`, …).
- **`client.Get/Post/Put/Patch/Delete`** reach endpoints outside the curated surface.

Optional collection fields in generated request params distinguish `nil` from an empty value: `nil` omits the field, while `[]string{}` or `map[string]any{}` sends an empty collection. The API decides whether that clears a value or is rejected.

## Nullable request fields

Use `bird.Value(value)` to send a value and `bird.Null[T]()` to send explicit null. The zero `Nullable[T]` omits an optional property. A required nullable property must be initialized: otherwise the call returns `*bird.RequestValidationError` with its `Field` path before sending, without retrying. For example, clear a number's inbound route with `VoiceInboundConfigurationPut{Route: bird.Null[bird.VoiceCallRouteWritable]()}`.

The following public fields change from pointers to `Nullable[T]`. Update struct literals to use `Value` or `Null`; test presence with `IsSpecified()`, test null with `IsNull()`, and read a non-null value with `MustGet()`.

| Field                                        | Before                              | After                                        |
| -------------------------------------------- | ----------------------------------- | -------------------------------------------- |
| `VoiceInboundConfigurationPut.Route`         | `*VoiceCallRouteWritable`           | `Nullable[VoiceCallRouteWritable]`           |
| `VoiceSequenceSavedExecutionCallSample.Orig` | `*VoiceSequenceSavedExecutionParty` | `Nullable[VoiceSequenceSavedExecutionParty]` |
| `VoiceSequenceSavedExecutionCallSample.Dest` | `*VoiceSequenceSavedExecutionParty` | `Nullable[VoiceSequenceSavedExecutionParty]` |
| `AMBListPickerItem.ImageIdentifier`          | `*string`                           | `Nullable[string]`                           |
| `AMBListPickerItem.Subtitle`                 | `*string`                           | `Nullable[string]`                           |
| `AMBListPickerSection.Order`                 | `*int`                              | `Nullable[int]`                              |

Apple Messages list-picker types are shared by requests and responses, so reading these three response fields also uses the nullable methods. Response-only fields keep their pointer types.

Untyped nullable request fields also carry explicit presence:

| Public parameter field                      | Before | After           |
| ------------------------------------------- | ------ | --------------- |
| `ContactPropertyCreateParams.FallbackValue` | `any`  | `Nullable[any]` |
| `ContactPropertyUpdateParams.FallbackValue` | `any`  | `Nullable[any]` |
| `RealtimePublishParams.Data`                | `any`  | `Nullable[any]` |
| `RealtimeBatchEventParams.Data`             | `any`  | `Nullable[any]` |
| `RealtimeMemberSendParams.Data`             | `any`  | `Nullable[any]` |

Use `bird.Value[any](value)` for a payload or fallback and `bird.Null[any]()` for JSON null. Realtime null delivers a null payload; contact-property update null removes its fallback.

`VoiceVerifiedNumbersUpdateParams.Name` retains `Nullable[string]` but now requires an explicit label or null.

Closed update objects now state fields already required by server validation:

| Public Go field                    | Before                       | After                       |
| ---------------------------------- | ---------------------------- | --------------------------- |
| `VoiceRingGroupsUpdateParams.Name` | `*string`                    | `string`                    |
| `VoiceTrunkMediaUpdate.Encryption` | `*VoiceTrunkMediaEncryption` | `VoiceTrunkMediaEncryption` |

For example, replace `VoiceRingGroupsUpdateParams{Name: bird.Ptr("Support")}` with `VoiceRingGroupsUpdateParams{Name: "Support"}`, and replace `VoiceTrunkMediaUpdate{Encryption: bird.Ptr(mode)}` with `VoiceTrunkMediaUpdate{Encryption: mode}`. `VoiceTrunksUpdateParams.Media` remains optional; only a supplied media object requires encryption. `AdminMarketingStillUpdate.Name` is staff-only and adds no public SDK field.

`VoiceDailySpendLimitUpdate.WorkspaceLimit` and `VoiceDestinationDailySpendLimitUpdate.Limit` already use `Nullable[Money]`; they now reject an unset value locally as required by the API.

## Broadcast template updates

`BroadcastsUpdateParams.Template` now accepts `Nullable[BroadcastsParameterTemplate]` instead of `Nullable[string]`. Move the replaced top-level `Language`, `Parameters`, and `ParameterOverrides` into that template. The wire body is unchanged; an explicitly supplied empty template is now expressible.

```go
templateID := "emt_…"
// Before
params := bird.BroadcastsUpdateParams{Template: bird.Value(templateID), Language: bird.Value("en")}
// After
params := bird.BroadcastsUpdateParams{Template: bird.Value(bird.BroadcastsParameterTemplate{
    ID: &templateID, Language: bird.Value("en"),
})}
// Sends {"template":{}}
empty := bird.BroadcastsUpdateParams{Template: bird.Value(bird.BroadcastsParameterTemplate{})}
```

## Examples

Runnable, per-method examples live in [`example_test.go`](./example_test.go) and render under each method on [pkg.go.dev](https://pkg.go.dev/github.com/messagebird/bird-sdk-go): sending (simple and rich), error handling, get, pagination, channel defaults, the webhook receiver, and the escape hatch.

## Design

The wire types and a low-level client are generated from the OpenAPI spec into `internal/oapi`; this package is the hand-written idiomatic layer on top.

## Apple Messages quickstart

Use `client.Amb.Send` to reply to an open, customer-initiated conversation as a configured, connected business account. Set `BIRD_API_KEY` and `AMB_CONVERSATION_ID`, then run the [Apple Messages example](examples/quickstart-amb/main.go). The key needs `amb:read`, `amb:write` and `amb_management:read`.

The example sends a real reply. Verify the recipient and content before running it. `accepted` means queued; `sent` means Apple gateway acceptance, not device delivery or a read receipt. Native payment, authentication and invitation requests are not part of this public channel release.

## Client enrichment

Requests can include caller provenance and **configured model (self-reported)** for matched Claude Code and Gemini environments. Only known public model IDs/aliases or `other` are sent. Set `BIRD_CLIENT_ENRICHMENT=0` to opt out; `DO_NOT_TRACK` and `BIRD_TELEMETRY=0` also disable enrichment. Existing header options accept a per-request `Bird-Model` declaration or `Bird-Enrichment: 0` opt-out; model declarations are normalized before sending.

## Media library

Browse assets, organize folders, and inspect permitted references with the media-library resources. Upload files through the [HTTP API or CLI](https://bird.com/docs/guides/media-library), then use their IDs with the SDK. Deleting a library entry leaves its public URL serving.

See the [media-library example](examples/quickstart-media-library/main.go).
