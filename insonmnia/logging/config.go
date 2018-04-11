package logging

// Config represents a logging config.
type Config struct {
	Level *Level `yaml:"level" required:"true" default:"info"`
}

func (m *Config) LogLevel() Level {
	return *m.Level
}
