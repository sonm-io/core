package logging

// Config represents a logging config.
type Config struct {
	Level Level `yaml:"level" required:"true" default:"info"`
}
