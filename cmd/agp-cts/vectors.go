package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openagp/cts/internal/canonical"
	"github.com/openagp/cts/internal/sign"
	"github.com/openagp/cts/internal/vectors"
)

func vectorsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vectors",
		Short: "Run the embedded cross-language test vectors against the Go implementation",
		Long: `Validate that this CTS binary's Go implementation produces byte-identical
canonicalization and signatures to the reference Python and TypeScript SDKs.

If this command exits non-zero, this build of agp-cts is non-conformant
with v0.1. Investigate before claiming any conformance level.`,
		RunE: runVectors,
	}
}

func runVectors(cmd *cobra.Command, args []string) error {
	canonOK, canonFail, err := runCanonicalizationVectors()
	if err != nil {
		return err
	}
	signOK, signFail, err := runSigningVectors()
	if err != nil {
		return err
	}

	policyVectors, err := vectors.PolicyVectors()
	if err != nil {
		return err
	}
	policyTotal := len(policyVectors)

	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════")
	fmt.Println(" AGP v0.1 conformance summary")
	fmt.Println("══════════════════════════════════════════════════")
	fmt.Printf("  RFC 8785 canonicalization:   %d/%d\n", canonOK, canonOK+canonFail)
	fmt.Printf("  Sign + verify (Ed25519):     %d/%d\n", signOK, signOK+signFail)
	fmt.Printf("  Policy DSL evaluation:       skipped (Go evaluator deferred to CTS v0.2)  [%d available]\n", policyTotal)
	fmt.Println()

	if canonFail > 0 || signFail > 0 {
		fmt.Println("  Result: \033[31mFAIL\033[0m — this implementation is NOT v0.1 conformant.")
		return fmt.Errorf("vectors failed")
	}
	fmt.Println("  Result: \033[32mPASS\033[0m — Go implementation is byte-identical to the reference.")
	return nil
}

func runCanonicalizationVectors() (ok, fail int, err error) {
	vs, err := vectors.CanonicalizationVectors()
	if err != nil {
		return 0, 0, err
	}
	fmt.Println("→ RFC 8785 canonicalization vectors")
	for _, v := range vs {
		got, err := canonical.Canonicalize(v.Input)
		if err != nil {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — canonicalize error: %v\n", v.Name, err)
			fail++
			continue
		}
		gotHex := hex.EncodeToString(got)
		if gotHex != v.ExpectedCanonicalUTF8Hex {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — bytes mismatch\n", v.Name)
			fmt.Printf("       got:      %s\n", gotHex)
			fmt.Printf("       expected: %s\n", v.ExpectedCanonicalUTF8Hex)
			fail++
			continue
		}
		// Also verify the SHA-256 hash, paranoia layer.
		h := sha256.Sum256(got)
		gotSHA := hex.EncodeToString(h[:])
		if gotSHA != v.ExpectedCanonicalSHA256Hex {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — sha256 mismatch (canonical bytes match!)\n", v.Name)
			fail++
			continue
		}
		fmt.Printf("  \033[32mok\033[0m    %s\n", v.Name)
		ok++
	}
	return ok, fail, nil
}

func runSigningVectors() (ok, fail int, err error) {
	vs, keypair, err := vectors.SigningVectors()
	if err != nil {
		return 0, 0, err
	}
	fmt.Println("\n→ Sign + verify vectors (deterministic Ed25519, fixed test key)")

	priv, err := sign.PrivateKeyFromBase64(keypair.PrivateKeyB64)
	if err != nil {
		return 0, 0, fmt.Errorf("decode test private key: %w", err)
	}
	pub, err := sign.PublicKeyFromBase64(keypair.PublicKeyB64)
	if err != nil {
		return 0, 0, fmt.Errorf("decode test public key: %w", err)
	}

	for _, v := range vs {
		// Deep-clone the input so successive vector runs don't pollute it.
		inputBytes, err := json.Marshal(v.Input)
		if err != nil {
			return ok, fail, err
		}
		var msg1 map[string]any
		_ = json.Unmarshal(inputBytes, &msg1)
		var msg2 map[string]any
		_ = json.Unmarshal(inputBytes, &msg2)

		// Check the signing input bytes match.
		input, err := sign.BuildSigningInput(msg1, v.KeyID)
		if err != nil {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — signing input error: %v\n", v.Name, err)
			fail++
			continue
		}
		gotHex := hex.EncodeToString(input)
		if gotHex != v.ExpectedSigningInputUTF8Hex {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — signing-input bytes mismatch\n", v.Name)
			fmt.Printf("       got:      %s\n", gotHex)
			fmt.Printf("       expected: %s\n", v.ExpectedSigningInputUTF8Hex)
			fail++
			continue
		}

		// Sign with the test key and check the signature matches (Ed25519 is deterministic).
		signed, err := sign.Sign(msg2, priv, v.KeyID)
		if err != nil {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — sign error: %v\n", v.Name, err)
			fail++
			continue
		}
		gotSig := signed["signature"].(map[string]any)["value"].(string)
		if gotSig != v.ExpectedSignatureB64 {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — signature mismatch\n", v.Name)
			fmt.Printf("       got:      %s\n", gotSig)
			fmt.Printf("       expected: %s\n", v.ExpectedSignatureB64)
			fail++
			continue
		}

		// Round-trip verify (proves the public key is the right counterpart).
		if err := sign.Verify(signed, pub); err != nil {
			fmt.Printf("  \033[31mFAIL\033[0m  %s — verify error: %v\n", v.Name, err)
			fail++
			continue
		}
		fmt.Printf("  \033[32mok\033[0m    %s\n", v.Name)
		ok++
	}
	return ok, fail, nil
}
