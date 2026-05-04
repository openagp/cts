package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/openagp/cts/internal/schema"
)

func validateCmd() *cobra.Command {
	var kind string

	cmd := &cobra.Command{
		Use:   "validate <file>",
		Short: "Schema-validate a JSON file against a bundled AGP schema",
		Long: `Validate a JSON file against the bundled v0.1 schema for the given message kind.

Example:
  agp-cts validate --kind event fixtures/events/01-tool-call-allowed.json

Available kinds:
  event, policy, decision-request, decision-response, discovery`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			validator, err := schema.NewValidator(kind)
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read %s: %w", args[0], err)
			}
			if err := validator.ValidateRaw(raw); err != nil {
				return err
			}
			fmt.Printf("OK  %s  (against bundled %s schema)\n", args[0], kind)
			return nil
		},
	}
	cmd.Flags().StringVar(&kind, "kind", "event",
		"message kind: event | policy | decision-request | decision-response | discovery")
	return cmd
}
