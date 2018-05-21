package npp

type rendezvousError struct {
	error
}

func newRendezvousError(err error) error {
	if err == nil {
		return nil
	}

	return &rendezvousError{err}
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

// Temporary returns true if this error is temporary.
//
// Used to trick into submission gRPC's machinery about exponentially delaying
// failed connections.
func (m *relayError) Temporary() bool {
	return true
}
