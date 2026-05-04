package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openagp/cts/internal/schema"
	"github.com/openagp/cts/internal/sign"
)

func verifyCmd() *cobra.Command {
	var publicKeyB64 string
	var kind string

	cmd := &cobra.Command{
		Use:   "verify <file>",
		Short: "Verify the Ed25519 signature on a signed AGP message",
		Long: `Verify a signed AGP message per ADR 0001.

Steps:
  1. Schema-validate the file against the bundled v0.1 schema for --kind.
  2. Reject if signature.alg != "Ed25519".
  3. Reconstruct the canonical signing input.
  4. Verify the signature against --public-key.

The public key MUST be supplied as a standard-base64-encoded 32-byte
Ed25519 raw public key (44 characters with padding).

Example:
  agp-cts verify --kind event \\
    --public-key IEEpcR0zvJKhUmabE+XidbzZsYo+Tat1sB1eQ4uWoqQ= \\
    output/event.signed.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if publicKeyB64 == "" {
				return fmt.Errorf("--public-key is required")
			}

			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read %s: %w", args[0], err)
			}

			validator, err := schema.NewValidator(kind)
			if err != nil {
				return err
			}
			if err := validator.ValidateRaw(raw); err != nil {
				return err
			}

			var msg map[string]any
			if err := json.Unmarshal(raw, &msg); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}

			pk, err := sign.PublicKeyFromBase64(publicKeyB64)
			if err != nil {
				return err
			}
			if err := sign.Verify(msg, pk); err != nil {
				return err
			}
			fmt.Printf("OK  %s  signature verifies (kind=%s)\n", args[0], kind)
			return nil
		},
	}
	cmd.Flags().StringVar(&publicKeyB64, "public-key", "",
		"Base64-encoded 32-byte Ed25519 public key (required)")
	cmd.Flags().StringVar(&kind, "kind", "event",
		"message kind: event | policy | decision-request | decision-response | discovery")
	return cmd
}
