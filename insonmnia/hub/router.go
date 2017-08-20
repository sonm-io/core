package hub

type router interface {
	RegisterRoute(ID string, protocol string, realIP string, realPort uint16) error
}

type directRouter struct {
}

func newDirectRouter() router {
	return &directRouter{}
}

func (r *directRouter) RegisterRoute(ID string, protocol string, realIP string, realPort uint16) error {
	return nil
}
