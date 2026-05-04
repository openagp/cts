package sign_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/openagp/cts/internal/sign"
	"github.com/openagp/cts/internal/vectors"
)

func TestSigningVectors(t *testing.T) {
	vs, keypair, err := vectors.SigningVectors()
	if err != nil {
		t.Fatalf("load vectors: %v", err)
	}
	if len(vs) == 0 {
		t.Fatal("no signing vectors found")
	}

	priv, err := sign.PrivateKeyFromBase64(keypair.PrivateKeyB64)
	if err != nil {
		t.Fatalf("decode test private key: %v", err)
	}
	pub, err := sign.PublicKeyFromBase64(keypair.PublicKeyB64)
	if err != nil {
		t.Fatalf("decode test public key: %v", err)
	}

	for _, v := range vs {
		t.Run(v.Name, func(t *testing.T) {
			// Deep clone the input so mutating BuildSigningInput doesn't bleed.
			b, _ := json.Marshal(v.Input)
			var msg1 map[string]any
			_ = json.Unmarshal(b, &msg1)
			var msg2 map[string]any
			_ = json.Unmarshal(b, &msg2)

			signingInput, err := sign.BuildSigningInput(msg1, v.KeyID)
			if err != nil {
				t.Fatalf("BuildSigningInput: %v", err)
			}
			if got := hex.EncodeToString(signingInput); got != v.ExpectedSigningInputUTF8Hex {
				t.Errorf("signing input bytes mismatch\n  got:      %s\n  expected: %s",
					got, v.ExpectedSigningInputUTF8Hex)
			}

			signed, err := sign.Sign(msg2, priv, v.KeyID)
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}
			gotSig := signed["signature"].(map[string]any)["value"].(string)
			if gotSig != v.ExpectedSignatureB64 {
				t.Errorf("signature bytes mismatch\n  got:      %s\n  expected: %s",
					gotSig, v.ExpectedSignatureB64)
			}

			if err := sign.Verify(signed, pub); err != nil {
				t.Errorf("verify: %v", err)
			}
		})
	}
}

func TestVerifyTamperedFieldFails(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	msg := map[string]any{
		"agp_version":     "0.1",
		"schema_version":  "1.0",
		"event_id":        "evt_x",
		"occurred_at":     "2026-08-12T14:23:11.412Z",
		"actor":           map[string]any{"vendor": "x", "agent_id": "y"},
		"action":          map[string]any{"type": "tool_call", "tool_name": "foo"},
	}
	signed, err := sign.Sign(msg, priv, "k1")
	if err != nil {
		t.Fatal(err)
	}

	// Tamper.
	signed["action"].(map[string]any)["tool_name"] = "evil.tool"

	if err := sign.Verify(signed, pub); err == nil {
		t.Fatal("expected verify to fail on tampered field")
	}
}

func TestVerifyWrongKeyFails(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	otherPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	msg := map[string]any{
		"agp_version":    "0.1",
		"schema_version": "1.0",
		"event_id":       "evt_x",
		"actor":          map[string]any{"vendor": "x", "agent_id": "y"},
		"action":         map[string]any{"type": "tool_call"},
	}
	signed, err := sign.Sign(msg, priv, "k1")
	if err != nil {
		t.Fatal(err)
	}

	if err := sign.Verify(signed, otherPub); err == nil {
		t.Fatal("expected verify to fail with wrong public key")
	}
}

func TestVerifyRejectsNonEd25519Alg(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pub, _, _ := ed25519.GenerateKey(rand.Reader)

	msg := map[string]any{
		"agp_version":    "0.1",
		"schema_version": "1.0",
		"event_id":       "evt_x",
		"actor":          map[string]any{"vendor": "x", "agent_id": "y"},
		"action":         map[string]any{"type": "tool_call"},
	}
	signed, err := sign.Sign(msg, priv, "k1")
	if err != nil {
		t.Fatal(err)
	}
	signed["signature"].(map[string]any)["alg"] = "RS256"

	if err := sign.Verify(signed, pub); err == nil {
		t.Fatal("expected verify to reject non-Ed25519 alg")
	}
}
