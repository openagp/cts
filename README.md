# openagp/cts

**AGP Conformance Test Suite.** A CLI for self-testing AGP implementations and producing signed conformance reports.

## Status

Scaffold. Implementation tracked in [§4.2 Phase 2](https://github.com/openagp/spec/blob/main/concept-and-spec.md#42-build-order--what-claude-code-should-build-first) of the spec.

## Planned commands

```bash
agp-cts validate-vendor --endpoint https://vendor.example.com/agp/v0/
agp-cts validate-plane  --endpoint https://plane.example.com/agp/v0/
agp-cts validate-fixture path/to/event.json
```

Output is a signed JSON conformance report listing the conformance level passed (L1, L2, or L3).

## Distribution

Single static binary for Linux, macOS, Windows. Releases via GitHub Releases.

## Language

Go (Cobra CLI scaffolding).

## License

[Apache-2.0](LICENSE).
