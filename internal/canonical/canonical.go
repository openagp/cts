// Package canonical wraps RFC 8785 JSON Canonicalization Scheme (JCS) for AGP.
//
// Per ADR 0001, AGP signs over RFC 8785 canonical bytes. This package is
// the only place in the CTS that knows how to produce those bytes; if the
// JCS implementation changes, only this file changes.
package canonical

import (
	"encoding/json"
	"fmt"

	"github.com/gowebpki/jcs"
)

// Canonicalize returns the RFC 8785 canonical UTF-8 bytes for an arbitrary
// Go value. The value MUST be JSON-serializable.
func Canonicalize(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: marshal failed: %w", err)
	}
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: jcs transform failed: %w", err)
	}
	return canonical, nil
}

// CanonicalizeRaw runs JCS over already-encoded JSON bytes. Useful for
// validating files on disk without re-decoding.
func CanonicalizeRaw(raw []byte) ([]byte, error) {
	canonical, err := jcs.Transform(raw)
	if err != nil {
		return nil, fmt.Errorf("canonicalize: jcs transform failed: %w", err)
	}
	return canonical, nil
}
