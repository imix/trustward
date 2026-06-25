package validate_test

import (
	"testing"

	"github.com/imix/trustward/internal/project"
	"github.com/imix/trustward/internal/validate"
)

// The shipped example model is the reference for what a correct project
// looks like — it must validate without issues.
func TestCheck_ExampleModelIsClean(t *testing.T) {
	proj, err := project.Load("../../../example/fire-protection-system")
	if err != nil {
		t.Fatalf("Load example: %v", err)
	}

	issues := validate.Check(proj)

	for _, is := range issues {
		t.Errorf("unexpected issue: %s", is)
	}
}
