// Package sign implements ADR 0001's signing protocol in Go.
//
// Mirrors openagp.events in Python and src/events.ts in TypeScript. Cross-
// language interop is verified against the test vectors committed in
// openagp/spec/test-vectors/.
package sign

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/openagp/cts/internal/canonical"
)

const SigAlg = "Ed25519"

// ErrInvalidSignature is returned when a signature does not verify or is
// missing required fields.
var ErrInvalidSignature = errors.New("invalid signature")

// BuildSigningInput constructs the canonical signing input per ADR 0001
// §"To sign" steps 1–2: set signature to {key_id, alg} (no value field),
// then JCS-canonicalize the entire message.
//
// `message` is mutated. Callers wanting to preserve the original should
// pass a copy.
func BuildSigningInput(message map[string]any, keyID string) ([]byte, error) {
	message["signature"] = map[string]any{
		"key_id": keyID,
		"alg":    SigAlg,
	}
	return canonical.Canonicalize(message)
}

// Sign signs a message in place per ADR 0001. Returns the message with
// signature.value populated. The input map IS mutated (Go has no
// inexpensive deep clone for arbitrary maps; callers wanting immutability
// should clone first).
func Sign(message map[string]any, privateKey ed25519.PrivateKey, keyID string) (map[string]any, error) {
	signingInput, err := BuildSigningInput(message, keyID)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(privateKey, signingInput)
	message["signature"] = map[string]any{
		"key_id": keyID,
		"alg":    SigAlg,
		"value":  base64.StdEncoding.EncodeToString(sig),
	}
	return message, nil
}

// Verify verifies a signed AGP message per ADR 0001 against `publicKey`.
// Returns nil on success, an error wrapping ErrInvalidSignature on failure.
func Verify(message map[string]any, publicKey ed25519.PublicKey) error {
	sigAny, ok := message["signature"]
	if !ok {
		return fmt.Errorf("%w: message has no signature field", ErrInvalidSignature)
	}
	sig, ok := sigAny.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: signature is not an object", ErrInvalidSignature)
	}

	alg, _ := sig["alg"].(string)
	if alg != SigAlg {
		return fmt.Errorf("%w: unsupported alg %q (v0.1 requires %q)", ErrInvalidSignature, alg, SigAlg)
	}

	keyID, _ := sig["key_id"].(string)
	if keyID == "" {
		return fmt.Errorf("%w: signature.key_id missing", ErrInvalidSignature)
	}

	valueB64, _ := sig["value"].(string)
	if valueB64 == "" {
		return fmt.Errorf("%w: signature.value missing", ErrInvalidSignature)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(valueB64)
	if err != nil {
		return fmt.Errorf("%w: signature.value not valid base64: %v", ErrInvalidSignature, err)
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("%w: expected %d-byte Ed25519 signature, got %d",
			ErrInvalidSignature, ed25519.SignatureSize, len(sigBytes))
	}

	// Reconstruct the signing input. We replace signature with {key_id,
	// alg} (dropping value) before canonicalizing.
	signingInput, err := BuildSigningInput(message, keyID)
	if err != nil {
		return err
	}
	// Restore the original signature object so the caller still has it.
	message["signature"] = sig

	if !ed25519.Verify(publicKey, signingInput, sigBytes) {
		return fmt.Errorf("%w: Ed25519 verification failed", ErrInvalidSignature)
	}
	return nil
}

// PublicKeyFromBase64 decodes a 32-byte Ed25519 public key from standard base64.
func PublicKeyFromBase64(b64 string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("public key: base64 decode failed: %w", err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key: expected %d bytes, got %d",
			ed25519.PublicKeySize, len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

// PrivateKeyFromBase64 decodes a 32-byte raw Ed25519 seed from base64 and
// expands it to the 64-byte form used by crypto/ed25519.
func PrivateKeyFromBase64(b64 string) (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("private key: base64 decode failed: %w", err)
	}
	if len(raw) != ed25519.SeedSize {
		return nil, fmt.Errorf("private key: expected %d-byte seed, got %d",
			ed25519.SeedSize, len(raw))
	}
	return ed25519.NewKeyFromSeed(raw), nil
}
