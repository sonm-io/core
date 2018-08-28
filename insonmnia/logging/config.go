package logging

import "os"

const (
	StdoutLogOutput = "stdout"
	StderrLogOutput = "stderr"
)

// Config represents a logging config.
type Config struct {
	Level  *Level `yaml:"level" required:"true" default:"info"`
	Output string `yaml:"output" default:"stdout"`
}

func (m *Config) LogLevel() Level {
	return *m.Level
}

func fdFromString(out string) *os.File {
	switch out {
	case StderrLogOutput:
		return os.Stderr
	case StdoutLogOutput:
		return os.Stdout
	default:
		return nil
	}
}
