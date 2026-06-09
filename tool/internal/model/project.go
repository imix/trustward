package model

// Project is the accumulated security model loaded from a directory.
// Any file in the import graph can contribute to any of these fields.
type Project struct {
	Version    Version
	SystemMeta *SystemMeta
	Assets     []Asset
	Components []Component
	TrustZones []TrustZone
	DataFlows  []DataFlow
	Threats    []Threat
	Controls   []Control
	Catalogs   []Catalog
}
