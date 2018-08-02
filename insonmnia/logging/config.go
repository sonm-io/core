package logging

// Config represents a logging config.
type Config struct {
	Level *Level `yaml:"level" default:"info"`
}

func (m *Config) LogLevel() Level {
	return *m.Level
}
