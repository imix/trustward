package validate_test

import (
	"strings"
	"testing"

	"github.com/imix/trustward/internal/model"
	"github.com/imix/trustward/internal/validate"
)

// issueMentioning reports whether some issue's text contains all wants.
func issueMentioning(issues []validate.Issue, wants ...string) bool {
	for _, is := range issues {
		text := is.String()
		found := true
		for _, w := range wants {
			if !strings.Contains(text, w) {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

func TestCheck_ReferenceNeedsVersion(t *testing.T) {
	p := &model.Project{
		References: []model.Reference{
			{ID: "variants", Title: "Variant register", Version: "1.0", Location: "./variants.md"},
			{ID: "reqs", Title: "Requirements", Location: "https://example.com/reqs"}, // no version
		},
	}
	issues := validate.Check(p)
	if !issueMentioning(issues, "reqs", "version") {
		t.Errorf("want missing-version issue for reqs, got %v", issues)
	}
	if issueMentioning(issues, "variants", "version") {
		t.Errorf("variants pins a version; should not be flagged: %v", issues)
	}
}

func TestCheck_RiskFieldsValidated(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true},
		Threats: []model.Threat{{
			ID:         "threat-x",
			Likelihood: "extreme", // out of scale
			Impact:     "high",
			Treatment:  "frobnicate", // not a valid treatment
			Owner:      "alice",
		}},
	}
	issues := validate.Check(p)
	if !issueMentioning(issues, "threat-x", "likelihood") {
		t.Errorf("want out-of-scale likelihood issue, got %v", issues)
	}
	if !issueMentioning(issues, "threat-x", "treatment") {
		t.Errorf("want bad-treatment issue, got %v", issues)
	}
}

func TestCheck_AttackFactorsValidated(t *testing.T) {
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "etsi-tvra", Accept: []string{"low"}, Set: true},
		Threats: []model.Threat{{
			ID:        "threat-x",
			Impact:    "high",
			Treatment: "mitigate", Owner: "alice",
			Attack: &model.AttackPotential{
				Expertise: "wizard", // not a valid factor value
				Knowledge: "public", Opportunity: "easy", Equipment: "standard",
			},
		}},
	}
	if !issueMentioning(validate.Check(p), "threat-x", "expertise") {
		t.Errorf("want invalid attack-factor issue, got %v", validate.Check(p))
	}
}

func TestCheck_UnknownRiskPolicyMethodRejected(t *testing.T) {
	// A typo'd method must be rejected, not silently scored as qualitative.
	p := &model.Project{
		RiskPolicy: model.RiskPolicy{Method: "etsi-tvra-typo", Set: true},
		Threats:    []model.Threat{{ID: "threat-x", Severity: "low"}},
	}
	if !issueMentioning(validate.Check(p), "risk-policy", "method") {
		t.Errorf("want unknown-method issue, got %v", validate.Check(p))
	}
}

func TestCheck_CRAGate(t *testing.T) {
	// risk-policy accepts only "low"; a high risk with no treatment is an open gap.
	base := model.RiskPolicy{Method: "qualitative", Accept: []string{"low"}, Set: true}
	untreated := &model.Project{
		RiskPolicy: base,
		Threats:    []model.Threat{{ID: "threat-x", Likelihood: "high", Impact: "high"}}, // -> critical
	}
	if !issueMentioning(validate.Check(untreated), "threat-x", "treatment") {
		t.Errorf("want CRA gate to flag unaccepted untreated risk, got %v", validate.Check(untreated))
	}

	treated := &model.Project{
		RiskPolicy: base,
		Threats: []model.Threat{{
			ID: "threat-x", Likelihood: "high", Impact: "high",
			Treatment: "mitigate", Owner: "alice",
		}},
	}
	if issueMentioning(validate.Check(treated), "threat-x", "treatment") {
		t.Errorf("treated+owned risk must pass the CRA gate, got %v", validate.Check(treated))
	}
}

func TestCheck_NoRiskPolicyNoGate(t *testing.T) {
	// Back-compat: without a risk-policy, untreated high-severity threats are fine.
	p := &model.Project{
		Threats: []model.Threat{{ID: "threat-x", Severity: "high"}},
	}
	if len(validate.Check(p)) != 0 {
		t.Errorf("no risk-policy should mean no risk gate, got %v", validate.Check(p))
	}
}

func TestCheck_ObjectiveRefsMustResolve(t *testing.T) {
	p := &model.Project{
		Objectives: []model.Objective{{ID: "obj-conf", Type: "confidentiality"}},
		Assets:     []model.Asset{{ID: "asset-a", Objectives: []string{"obj-conf", "obj-missing"}}},
		Threats:    []model.Threat{{ID: "threat-a", Violates: []string{"obj-gone"}}},
	}

	issues := validate.Check(p)

	if !issueMentioning(issues, "asset", "obj-missing") {
		t.Errorf("want unresolved asset objective issue, got %v", issues)
	}
	if !issueMentioning(issues, "threat-a", "obj-gone") {
		t.Errorf("want unresolved threat violates issue, got %v", issues)
	}
}

func TestCheck_ObjectiveTypeMustBeInCIAScale(t *testing.T) {
	p := &model.Project{
		Objectives: []model.Objective{
			{ID: "obj-ok", Type: "integrity"},
			{ID: "obj-bad", Type: "speed"},
		},
	}

	issues := validate.Check(p)

	if !issueMentioning(issues, "objective", "speed") {
		t.Errorf("want invalid objective type issue, got %v", issues)
	}
	if issueMentioning(issues, "obj-ok", "type") {
		t.Errorf("valid CIA type must not be flagged, got %v", issues)
	}
}

func TestCheck_ThreatMitigationMustMatchControl(t *testing.T) {
	p := &model.Project{
		Controls: []model.Control{{ID: "ctrl-a"}},
		Threats: []model.Threat{{
			ID:          "threat-x",
			Mitigations: []string{"ctrl-a", "ctrl-missing"},
		}},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "threat-x", "ctrl-missing") {
		t.Errorf("issue should name the threat and the unknown control, got %q", issues[0].String())
	}
}

func TestCheck_ThreatTargetMustMatchComponentOrFlow(t *testing.T) {
	p := &model.Project{
		Components: []model.Component{{ID: "comp-a"}},
		DataFlows:  []model.DataFlow{{ID: "flow-a", Connects: []string{"comp-a", "comp-a"}}},
		Threats: []model.Threat{
			{ID: "threat-on-comp", Target: "comp-a"},
			{ID: "threat-on-flow", Target: "flow-a"},
			{ID: "threat-dangling", Target: "nonexistent"},
		},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "threat-dangling", "nonexistent") {
		t.Errorf("issue should name the threat and the unknown target, got %q", issues[0].String())
	}
}

func TestCheck_ThreatAssetMustMatchAsset(t *testing.T) {
	p := &model.Project{
		Assets: []model.Asset{{ID: "asset-a"}},
		Threats: []model.Threat{
			{ID: "threat-ok", Asset: "asset-a"},
			{ID: "threat-no-asset"}, // empty asset is not an error
			{ID: "threat-bad", Asset: "asset-missing"},
		},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "threat-bad", "asset-missing") {
		t.Errorf("issue should name the threat and the unknown asset, got %q", issues[0].String())
	}
}

func TestCheck_ThreatRefMustMatchCatalogPattern(t *testing.T) {
	p := &model.Project{
		ThreatCatalogs: []model.ThreatCatalog{{
			ID:       "stride-ot",
			Patterns: []model.ThreatPattern{{ID: "spoof-field-device"}},
		}},
		Threats: []model.Threat{
			{ID: "threat-ok", Ref: "stride-ot::spoof-field-device"},
			{ID: "threat-bad-pattern", Ref: "stride-ot::no-such-pattern"},
			{ID: "threat-bad-catalog", Ref: "no-such-catalog::spoof-field-device"},
		},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "threat-bad-pattern", "stride-ot::no-such-pattern") {
		t.Errorf("missing issue for unresolved pattern, got %v", issues)
	}
	if !issueMentioning(issues, "threat-bad-catalog", "no-such-catalog::spoof-field-device") {
		t.Errorf("missing issue for unresolved catalog, got %v", issues)
	}
}

func TestCheck_ComponentRefsMustResolve(t *testing.T) {
	p := &model.Project{
		Assets:   []model.Asset{{ID: "asset-a"}},
		Controls: []model.Control{{ID: "ctrl-a"}},
		Components: []model.Component{{
			ID:       "comp-x",
			Assets:   []string{"asset-a", "asset-missing"},
			Controls: []string{"ctrl-a", "ctrl-missing"},
		}},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "comp-x", "asset-missing") {
		t.Errorf("missing issue for unknown asset, got %v", issues)
	}
	if !issueMentioning(issues, "comp-x", "ctrl-missing") {
		t.Errorf("missing issue for unknown control, got %v", issues)
	}
}

func TestCheck_TrustZoneMembersMustMatchComponents(t *testing.T) {
	p := &model.Project{
		Components: []model.Component{{ID: "comp-a"}},
		TrustZones: []model.TrustZone{{
			ID:      "zone-x",
			Members: []string{"comp-a", "comp-missing"},
		}},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "zone-x", "comp-missing") {
		t.Errorf("issue should name the zone and the unknown member, got %q", issues[0].String())
	}
}

func TestCheck_DataFlowRefsMustResolve(t *testing.T) {
	p := &model.Project{
		Assets:     []model.Asset{{ID: "asset-a"}},
		Components: []model.Component{{ID: "comp-a"}},
		DataFlows: []model.DataFlow{{
			ID:       "flow-x",
			Connects: []string{"comp-a", "comp-missing"},
			Assets:   []string{"asset-a", "asset-missing"},
		}},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "flow-x", "comp-missing") {
		t.Errorf("missing issue for unknown endpoint, got %v", issues)
	}
	if !issueMentioning(issues, "flow-x", "asset-missing") {
		t.Errorf("missing issue for unknown asset, got %v", issues)
	}
}

func TestCheck_DataFlowMustConnectExactlyTwoComponents(t *testing.T) {
	p := &model.Project{
		Components: []model.Component{{ID: "comp-a"}, {ID: "comp-b"}, {ID: "comp-c"}},
		DataFlows: []model.DataFlow{
			{ID: "flow-ok", Connects: []string{"comp-a", "comp-b"}},
			{ID: "flow-one-end", Connects: []string{"comp-a"}},
			{ID: "flow-three-ends", Connects: []string{"comp-a", "comp-b", "comp-c"}},
		},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "flow-one-end") {
		t.Errorf("missing issue for one-ended flow, got %v", issues)
	}
	if !issueMentioning(issues, "flow-three-ends") {
		t.Errorf("missing issue for three-ended flow, got %v", issues)
	}
}

func TestCheck_ControlRefMustMatchCatalogRequirement(t *testing.T) {
	p := &model.Project{
		Catalogs: []model.ControlCatalog{{
			ID:           "company-baseline",
			Requirements: []model.Requirement{{ID: "req-iam"}},
		}},
		Controls: []model.Control{
			{ID: "ctrl-ok", Ref: "company-baseline::req-iam"},
			{ID: "ctrl-no-ref"}, // a control without a catalog ref is not an error
			{ID: "ctrl-bad", Ref: "company-baseline::req-missing"},
		},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "ctrl-bad", "company-baseline::req-missing") {
		t.Errorf("issue should name the control and the unknown requirement, got %q", issues[0].String())
	}
}

func TestCheck_MissingIDIsReported(t *testing.T) {
	p := &model.Project{
		Assets: []model.Asset{{ID: "asset-a"}, {ID: ""}},
	}

	issues := validate.Check(p)

	if len(issues) != 1 {
		t.Fatalf("want exactly 1 issue, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "asset", "missing id") {
		t.Errorf("issue should flag the asset with no id, got %q", issues[0].String())
	}
}

func TestCheck_MultipleMissingIDsAreNotReportedAsDuplicates(t *testing.T) {
	p := &model.Project{
		Components: []model.Component{{ID: ""}, {ID: ""}},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	for _, is := range issues {
		if strings.Contains(is.String(), "duplicate") {
			t.Errorf("empty ids should read as missing, not duplicate, got %q", is.String())
		}
	}
}

func TestCheck_DuplicateIDsAreReported(t *testing.T) {
	p := &model.Project{
		Components: []model.Component{{ID: "comp-dup"}, {ID: "comp-dup"}},
		Threats:    []model.Threat{{ID: "threat-dup"}, {ID: "threat-dup"}},
	}

	issues := validate.Check(p)

	if len(issues) != 2 {
		t.Fatalf("want exactly 2 issues, got %d: %v", len(issues), issues)
	}
	if !issueMentioning(issues, "comp-dup", "duplicate") {
		t.Errorf("missing issue for duplicate component ID, got %v", issues)
	}
	if !issueMentioning(issues, "threat-dup", "duplicate") {
		t.Errorf("missing issue for duplicate threat ID, got %v", issues)
	}
}
