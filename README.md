# openagp/cts — Conformance Test Suite

**`agp-cts` is the official conformance tool for the Agent Governance Protocol.**

A single static binary. No runtime dependencies. Validates schemas, canonicalizes JSON to RFC 8785 bytes, verifies Ed25519 signatures per ADR 0001, and runs the cross-language test vectors that every conformant implementation must pass.

[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg?style=flat-square)](LICENSE)
[![Spec](https://img.shields.io/badge/spec-v0.1%20draft-blue.svg?style=flat-square)](https://github.com/openagp/spec/blob/main/concept-and-spec.md)
[![Go](https://img.shields.io/badge/go-1.22%2B-007d9c.svg?style=flat-square)](https://go.dev/)

## Install

```bash
# from source
make build && ./agp-cts version

# or grab a release binary from the Releases page (Linux/macOS/Windows × amd64/arm64)
```

## Usage

```bash
# Run the embedded cross-language test vectors against this binary's Go implementation.
# A passing run means agp-cts is byte-for-byte compatible with the reference Python and TypeScript SDKs.
agp-cts vectors

# Schema-validate a single file.
agp-cts validate --kind event path/to/event.json
agp-cts validate --kind policy path/to/policy.json

# Print the RFC 8785 canonical bytes of a JSON file.
agp-cts canonicalize path/to/file.json
agp-cts canonicalize --sha256 path/to/file.json

# Verify a signed AGP message.
agp-cts verify --kind event \
  --public-key IEEpcR0zvJKhUmabE+XidbzZsYo+Tat1sB1eQ4uWoqQ= \
  output/event.signed.json
```

`agp-cts vectors` is the most important command. It loads the test vectors embedded at build time and runs them through this binary's Go implementation. If the output bytes match the expected bytes from the spec — across canonicalization and signing — the implementation is conformant.

## What's embedded

At build time, `agp-cts` ships with:

- All v0.1 JSON Schemas (`event`, `policy`, `decision-request`, `decision-response`, `discovery`, `common`).
- All v0.1 cross-language test vectors (`v0.1-canonicalization.json`, `v0.1-signatures.json`, `v0.1-policy-decisions.json`).

Source: [`openagp/spec`](https://github.com/openagp/spec). Synchronized via `scripts/sync-spec.sh`; CI fails if embedded copies drift.

## What's not yet implemented

- **`validate-vendor --endpoint <url>`** — black-box conformance probe over HTTPS. The endpoint protocol is being formalized; HTTP probing lands in CTS v0.2.
- **`validate-plane --endpoint <url>`** — same, for plane-side endpoints.
- **`run-implementation --sign-script ... --verify-script ...`** — feed test vectors through external commands so any-language SDK can self-test by writing two wrapper scripts.
- **Policy DSL evaluator in Go.** v0.1 of the CTS includes the policy decision vectors but does not run them — the Python and TypeScript SDKs already test their evaluators against the same vectors. A Go evaluator lands in CTS v0.2 alongside `validate-vendor`.

## Cross-language guarantee

Three languages, same bytes:

| Implementation | Test count (vectors-only) |
|---|---|
| [`openagp/sdk-python`](https://github.com/openagp/sdk-python) | 11 |
| [`openagp/sdk-typescript`](https://github.com/openagp/sdk-typescript) | 11 |
| `openagp/cts` (this repo)                                       | 9 unit + 9 vector |

Every (input, expected canonical bytes) pair matches across all three. Every (input, expected Ed25519 signature) pair matches across all three. Differential testing as protocol-as-spec.

## Building

```bash
make build      # host platform
make release    # cross-compile for linux/darwin/windows × amd64/arm64
make test       # go test ./...
make vectors    # build + run the embedded vectors via the CLI
make sync-spec  # pull schemas/vectors from sibling ../spec
```

## Layout

```
cts/
├── cmd/agp-cts/                — CLI entry point + cobra subcommands
├── internal/canonical/         — RFC 8785 wrapper (gowebpki/jcs)
├── internal/sign/              — Ed25519 sign/verify per ADR 0001
├── internal/schema/            — Draft 2020-12 validator (santhosh-tekuri/jsonschema)
│   └── embed/                  — JSON Schemas (synced from spec)
├── internal/vectors/           — test vector loaders
│   └── embed/                  — test vectors (synced from spec)
└── scripts/sync-spec.sh        — sync from sibling ../spec
```

## License

[Apache-2.0](LICENSE).
