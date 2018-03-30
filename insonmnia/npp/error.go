package npp

// TransportError is an error that is returned when the underlying gRPC
// transport is broken.
type TransportError struct {
	error
}

type relayError struct {
	error
}

func newRelayError(err error) error {
	if err == nil {
		return nil
	}

	return &relayError{err}
}

func (m *relayError) Error() string {
	if m.error == nil {
		return "no error"
	}
	return m.error.Error()
}

// Temporary returns true if this error is temporary.
//
// Used to trick into submission gRPC's machinery about exponentially delaying
// failed connections.
func (m *relayError) Temporary() bool {
	return true
}
