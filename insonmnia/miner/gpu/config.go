package gpu

// Config contains options related to NVIDIA GPU support
type Config struct {
	Type string `yaml:"type"`
	Args Args   `yaml:"args"`
}

// Args are mode dependant options
type Args map[string]interface{}
