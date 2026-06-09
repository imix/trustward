package model

type Control struct {
	ID          string   `yaml:"id"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Evidence    []string `yaml:"evidence"`
}
