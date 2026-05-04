package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openagp/cts/internal/canonical"
)

func canonicalizeCmd() *cobra.Command {
	var hashOnly bool

	cmd := &cobra.Command{
		Use:   "canonicalize <file>",
		Short: "Print the RFC 8785 canonical bytes of a JSON file",
		Long: `Compute and print the RFC 8785 (JCS) canonical UTF-8 bytes of a JSON file.

This is the cross-language byte sequence that every conformant AGP
implementation MUST produce for a given input. Use this to debug
canonicalization differences between implementations.

By default, prints the canonical JSON. With --sha256, prints the hex
SHA-256 of the canonical bytes — useful for log lines or comparing
without dumping the full payload.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read %s: %w", args[0], err)
			}
			canon, err := canonical.CanonicalizeRaw(raw)
			if err != nil {
				return err
			}
			if hashOnly {
				h := sha256.Sum256(canon)
				fmt.Println(hex.EncodeToString(h[:]))
				return nil
			}
			os.Stdout.Write(canon)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().BoolVar(&hashOnly, "sha256", false, "print hex SHA-256 of the canonical bytes")
	return cmd
}
