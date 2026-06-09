package model

type Requirement struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Satisfies   []string `yaml:"satisfies"` // "catalog-id::req-id" references to other catalogs
}

type Catalog struct {
	ID           string        `yaml:"id"`
	Title        string        `yaml:"title"`
	Requirements []Requirement `yaml:"requirements"`
}

type Control struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Ref         string   `yaml:"ref"`      // optional: "catalog-id::req-id"
	Evidence    []string `yaml:"evidence"`
}
