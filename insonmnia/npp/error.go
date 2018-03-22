package npp

// TransportError is an error that is returned when the underlying gRPC
// transport is broken.
type TransportError struct {
	error
}

type RelayError struct {
	error
}

func (m *RelayError) Error() string {
	return m.error.Error()
}

func (m *RelayError) Temporary() bool {
	return true
}
