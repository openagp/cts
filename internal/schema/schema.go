// Package schema validates AGP messages against the bundled v0.1 JSON
// Schemas (Draft 2020-12). Schemas are embedded at build time so the CTS
// ships as a single static binary with no external runtime dependencies.
package schema

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed all:embed
var schemasFS embed.FS

// Kinds are the message kinds the CTS knows how to validate.
var Kinds = []string{
	"event",
	"policy",
	"decision-request",
	"decision-response",
	"discovery",
}

// Validator is a compiled validator for a specific message kind.
type Validator struct {
	kind   string
	schema *jsonschema.Schema
}

// NewValidator compiles the schema for `kind` from the embedded files,
// resolving `common.json` references against the bundled common schema.
func NewValidator(kind string) (*Validator, error) {
	if !validKind(kind) {
		return nil, fmt.Errorf("unknown kind %q; expected one of %v", kind, Kinds)
	}
	c := jsonschema.NewCompiler()

	// Load every embedded schema and register it under both its filename
	// and its $id (so $ref="common.json#..." and $ref="https://openagp.io/..." both resolve).
	entries, err := schemasFS.ReadDir("embed")
	if err != nil {
		return nil, fmt.Errorf("read embedded schemas: %w", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".json") {
			continue
		}
		raw, err := schemasFS.ReadFile("embed/" + ent.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", ent.Name(), err)
		}
		var doc any
		if err := json.Unmarshal(raw, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", ent.Name(), err)
		}
		// Register under bare filename (used by sibling-relative $ref).
		if err := c.AddResource(ent.Name(), doc); err != nil {
			return nil, fmt.Errorf("add %s: %w", ent.Name(), err)
		}
		// And under $id if present.
		if m, ok := doc.(map[string]any); ok {
			if id, ok := m["$id"].(string); ok && id != "" {
				_ = c.AddResource(id, doc)
			}
		}
	}

	sch, err := c.Compile(kind + ".json")
	if err != nil {
		return nil, fmt.Errorf("compile %s schema: %w", kind, err)
	}
	return &Validator{kind: kind, schema: sch}, nil
}

// Validate runs the compiled schema over `instance`. Returns nil if the
// instance conforms; otherwise an error listing the validation problems.
func (v *Validator) Validate(instance any) error {
	if err := v.schema.Validate(instance); err != nil {
		return fmt.Errorf("%s schema validation failed: %w", v.kind, err)
	}
	return nil
}

// ValidateRaw is a convenience wrapper that parses `raw` JSON before validating.
func (v *Validator) ValidateRaw(raw []byte) error {
	var instance any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&instance); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return v.Validate(instance)
}

// SchemasFS exposes the embedded schemas FS for callers that need to read
// schema files directly (e.g. for piping to other tools or printing).
func SchemasFS() embed.FS { return schemasFS }

func validKind(k string) bool {
	for _, x := range Kinds {
		if x == k {
			return true
		}
	}
	return false
}
