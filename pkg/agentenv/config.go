package agentenv

type Config struct {
	Project   Project             `yaml:"project"`
	Providers map[string]Provider `yaml:"providers"`
}

type Project struct {
	Name        string   `yaml:"name"`
	Language    string   `yaml:"language"`
	Description string   `yaml:"description"`
	Metadata    Metadata `yaml:"metadata"`
}

type Metadata struct {
	Repository string `yaml:"repository"`
}

type Provider struct {
	Template string `yaml:"template"`
	Output   string `yaml:"output"`
}
