// agp-cts — Conformance Test Suite for the Agent Governance Protocol.
//
// Single static binary. No external runtime deps. Schemas and test vectors
// are embedded at build time.
//
//   agp-cts validate --kind event path/to/event.json
//   agp-cts canonicalize path/to/file.json
//   agp-cts verify path/to/event.json --public-key <b64>
//   agp-cts vectors
//   agp-cts version
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	cliName    = "agp-cts"
	cliVersion = "0.0.0"
	agpVersion = "0.1"
)

func main() {
	root := &cobra.Command{
		Use:   cliName,
		Short: "Agent Governance Protocol Conformance Test Suite",
		Long: `agp-cts validates AGP implementations against the v0.1 specification.

This binary embeds the canonical JSON Schemas and the cross-language test
vectors from openagp/spec, and runs them against your implementation or
your own JSON files. A passing run means your implementation is byte-for-
byte compatible with the reference Python and TypeScript SDKs.`,
		SilenceUsage: true,
	}

	root.AddCommand(
		validateCmd(),
		canonicalizeCmd(),
		verifyCmd(),
		vectorsCmd(),
		versionCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print agp-cts and supported AGP version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s %s   (AGP v%s)\n", cliName, cliVersion, agpVersion)
		},
	}
}
