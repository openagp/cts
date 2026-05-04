package canonical_test

import (
	"encoding/hex"
	"testing"

	"github.com/openagp/cts/internal/canonical"
	"github.com/openagp/cts/internal/vectors"
)

func TestCanonicalizationVectors(t *testing.T) {
	vs, err := vectors.CanonicalizationVectors()
	if err != nil {
		t.Fatalf("load vectors: %v", err)
	}
	if len(vs) == 0 {
		t.Fatal("no canonicalization vectors found")
	}

	for _, v := range vs {
		t.Run(v.Name, func(t *testing.T) {
			got, err := canonical.Canonicalize(v.Input)
			if err != nil {
				t.Fatalf("canonicalize: %v", err)
			}
			gotHex := hex.EncodeToString(got)
			if gotHex != v.ExpectedCanonicalUTF8Hex {
				t.Errorf("canonical bytes mismatch\n  got:      %s\n  expected: %s",
					gotHex, v.ExpectedCanonicalUTF8Hex)
			}
		})
	}
}

func TestCanonicalizeKeyOrderIndependent(t *testing.T) {
	a := map[string]any{"b": 2, "a": 1, "c": []int{3, 2, 1}}
	b := map[string]any{"a": 1, "c": []int{3, 2, 1}, "b": 2}
	ca, err := canonical.Canonicalize(a)
	if err != nil {
		t.Fatal(err)
	}
	cb, err := canonical.Canonicalize(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(ca) != string(cb) {
		t.Errorf("expected identical canonical bytes\n  a: %s\n  b: %s", ca, cb)
	}
}
