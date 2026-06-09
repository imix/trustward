package model

type CompanyFile struct {
	Version  Version   `yaml:"version"`
	Controls []Control `yaml:"controls"`
}

type Control struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
