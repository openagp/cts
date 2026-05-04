package schema_test

import (
	"path/filepath"
	"testing"

	"os"

	"github.com/openagp/cts/internal/schema"
)

// All shipped fixtures must validate against their respective schemas.
// Path: ../../../spec/fixtures/{events,policies}/...
const fixturesEventsDir = "../../../spec/fixtures/events"
const fixturesPoliciesDir = "../../../spec/fixtures/policies"

func TestEventFixturesValidate(t *testing.T) {
	v, err := schema.NewValidator("event")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(fixturesEventsDir, "[0-9][0-9]-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("no event fixtures found at " + fixturesEventsDir + " — skipping (run from a checkout with sibling spec)")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if err := v.ValidateRaw(raw); err != nil {
				t.Errorf("validate %s: %v", f, err)
			}
		})
	}
}

func TestPolicyFixturesValidate(t *testing.T) {
	v, err := schema.NewValidator("policy")
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(fixturesPoliciesDir, "[0-9][0-9]-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("no policy fixtures found")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			raw, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			if err := v.ValidateRaw(raw); err != nil {
				t.Errorf("validate %s: %v", f, err)
			}
		})
	}
}

func TestUnknownKindRejected(t *testing.T) {
	if _, err := schema.NewValidator("nonexistent"); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}
