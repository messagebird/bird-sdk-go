# Contributing

Thanks for your interest in the Bird Go SDK.

This repository is a **read-only release mirror**. The SDK is generated and hand-maintained inside Bird's internal monorepo, which is the single source of truth; each release is published here as a snapshot. Commits are never authored directly on this mirror.

## Issues

Issues are welcome here — bug reports, missing functionality, and questions all help. Please include the SDK version (the release tag), your Go version, and a minimal reproduction where possible.

## Pull requests

Pull requests are welcome and reviewed here. Because this mirror is generated, an accepted change is **ported into the monorepo** and ships in the next release rather than being merged on the mirror directly. When that happens your commit is carried over with a `Co-authored-by` trailer so you keep authorship credit.

A few things to know before you open a PR:

- **Changes to the generated layer (`internal/oapi/`) cannot be accepted.** That code is regenerated from Bird's public OpenAPI specification on every release, so edits there are overwritten. If something is wrong in the generated output, open an issue describing the API behavior instead — the fix belongs in the spec or the generator config.
- The hand-written facade (`bird.go`, `email.go`, `webhooks.go`, and the rest of package `bird`) is the place for idiomatic improvements, bug fixes, and documentation.

## License

By contributing you agree that your contributions are licensed under the [MIT License](./LICENSE).
