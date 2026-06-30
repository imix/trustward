// Package schema validates trustward model files against the embedded JSON
// Schema — the same schema/trustward.schema.json that editors consume via
// yaml-language-server. Wiring it into `validate` means CI and the editor
// enforce one structural contract, with no second copy to drift.
//
// It checks structure and the closed vocabularies (severity, treatment,
// objective type, attack factors, risk-policy method, …). Referential
// integrity and the CRA gate stay in internal/validate; the two are
// complementary, not a replacement.
package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

//go:embed trustward.schema.json
var raw []byte

// Check validates every file in the model's import graph (from dir/system.yaml)
// against the schema and returns one message per violation. Read/parse errors
// are returned as messages too, so it never aborts the caller — the validate
// command folds these into its issue count alongside the referential checks.
func Check(dir string) []string {
	sch, err := jsonschema.CompileString("trustward.schema.json", string(raw))
	if err != nil {
		return []string{fmt.Sprintf("schema: %v", err)} // embedded schema is broken — a build defect
	}
	var msgs []string
	walk(filepath.Join(dir, "system.yaml"), map[string]bool{}, sch, &msgs)
	return msgs
}

// walk follows imports depth-first, validating each file.
// ponytail: a parse-light twin of project.loadGraph's import-following — kept
// separate so a structural error surfaces as a clear schema message instead of
// a Decode failure mid-merge. The two walkers are small; merging them would
// couple structural validation to the merge logic.
func walk(path string, visited map[string]bool, sch *jsonschema.Schema, msgs *[]string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		*msgs = append(*msgs, fmt.Sprintf("%s: %v", path, err))
		return
	}
	if visited[abs] {
		return
	}
	visited[abs] = true
	name := filepath.Base(abs)

	data, err := os.ReadFile(abs)
	if err != nil {
		*msgs = append(*msgs, fmt.Sprintf("%s: %v", name, err))
		return
	}

	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		*msgs = append(*msgs, fmt.Sprintf("%s: parsing: %v", name, err))
		return
	}
	// YAML yields Go types the validator doesn't speak (int, time.Time); a JSON
	// round-trip normalizes them to the JSON types jsonschema expects.
	if jb, err := json.Marshal(doc); err != nil {
		*msgs = append(*msgs, fmt.Sprintf("%s: %v", name, err))
	} else {
		var norm any
		_ = json.Unmarshal(jb, &norm)
		if err := sch.Validate(norm); err != nil {
			if ve, ok := err.(*jsonschema.ValidationError); ok {
				for _, line := range leaves(ve) {
					*msgs = append(*msgs, name+": "+line)
				}
			} else {
				*msgs = append(*msgs, name+": "+err.Error())
			}
		}
	}

	// Follow imports (path only — the rest is the loader's job).
	var top struct {
		Imports []struct {
			Path string `yaml:"path"`
		} `yaml:"imports"`
	}
	if err := yaml.Unmarshal(data, &top); err == nil {
		for _, imp := range top.Imports {
			walk(filepath.Join(filepath.Dir(abs), imp.Path), visited, sch, msgs)
		}
	}
}

// leaves flattens a jsonschema ValidationError to its leaf causes — the
// specific failures — as "instance/location: message" lines.
func leaves(ve *jsonschema.ValidationError) []string {
	if len(ve.Causes) == 0 {
		loc := strings.TrimPrefix(ve.InstanceLocation, "/")
		if loc == "" {
			loc = "(root)"
		}
		return []string{loc + ": " + ve.Message}
	}
	var out []string
	for _, c := range ve.Causes {
		out = append(out, leaves(c)...)
	}
	return out
}
