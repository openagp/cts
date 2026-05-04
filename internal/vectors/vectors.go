// Package vectors loads the cross-language test vectors that every
// conformant AGP implementation must pass.
package vectors

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed all:embed
var vectorsFS embed.FS

// CanonVector represents one RFC 8785 canonicalization test case.
type CanonVector struct {
	Name                       string `json:"name"`
	Description                string `json:"description,omitempty"`
	Input                      any    `json:"input"`
	ExpectedCanonicalUTF8Hex   string `json:"expected_canonical_utf8_hex"`
	ExpectedCanonicalSHA256Hex string `json:"expected_canonical_sha256_hex"`
}

// SignVector represents one signing test case (input + expected signature
// for a fixed test keypair).
type SignVector struct {
	Name                        string         `json:"name"`
	Input                       map[string]any `json:"input"`
	KeyID                       string         `json:"key_id"`
	ExpectedSigningInputUTF8Hex string         `json:"expected_signing_input_utf8_hex"`
	ExpectedSignatureB64        string         `json:"expected_signature_b64"`
}

// PolicyVector represents one policy decision test case.
type PolicyVector struct {
	Name             string         `json:"name"`
	PolicyFixture    string         `json:"policy_fixture"`
	Event            map[string]any `json:"event"`
	ExpectedDecision string         `json:"expected_decision"`
	ExpectedRuleID   string         `json:"expected_rule_id"`
	ExpectedAnnotate map[string]any `json:"expected_annotate"`
}

// CanonicalizationVectors loads the v0.1 RFC 8785 vectors.
func CanonicalizationVectors() ([]CanonVector, error) {
	data, err := vectorsFS.ReadFile("embed/v0.1-canonicalization.json")
	if err != nil {
		return nil, fmt.Errorf("load canonicalization vectors: %w", err)
	}
	var doc struct {
		Vectors []CanonVector `json:"vectors"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc.Vectors, nil
}

// SigningTestKeypair is the deterministic test keypair embedded in
// v0.1-signatures.json. Same in every language.
type SigningTestKeypair struct {
	KeyID         string `json:"key_id"`
	PrivateKeyB64 string `json:"private_key_b64"`
	PublicKeyB64  string `json:"public_key_b64"`
}

// SigningVectors loads the v0.1 signing vectors and returns them with the
// shared test keypair.
func SigningVectors() ([]SignVector, SigningTestKeypair, error) {
	data, err := vectorsFS.ReadFile("embed/v0.1-signatures.json")
	if err != nil {
		return nil, SigningTestKeypair{}, fmt.Errorf("load signing vectors: %w", err)
	}
	var doc struct {
		TestKeypair SigningTestKeypair `json:"test_keypair"`
		Vectors     []SignVector       `json:"vectors"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, SigningTestKeypair{}, err
	}
	return doc.Vectors, doc.TestKeypair, nil
}

// PolicyVectors loads the v0.1 policy decision vectors. The CTS does not
// (yet) include a Go policy evaluator, so these are loaded for inventory
// only — the policy evaluator runtime check is deferred to v0.2 of the CTS.
func PolicyVectors() ([]PolicyVector, error) {
	data, err := vectorsFS.ReadFile("embed/v0.1-policy-decisions.json")
	if err != nil {
		// Non-fatal: policy vectors are optional in v0.1.
		return nil, nil
	}
	var doc struct {
		Vectors []PolicyVector `json:"vectors"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc.Vectors, nil
}
