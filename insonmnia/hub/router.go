package hub

type route struct {
	ID          string
	Protocol    string
	Host        string
	Port        uint16
	BackendHost string
	BackendPort uint16
}

type router interface {
	RegisterRoute(ID string, protocol string, realIP string, realPort uint16) (*route, error)
}

type directRouter struct {
}

func newDirectRouter() router {
	return &directRouter{}
}

func (r *directRouter) RegisterRoute(ID string, protocol string, realIP string, realPort uint16) (*route, error) {
	route := &route{
		ID:          ID,
		Protocol:    protocol,
		Host:        realIP,
		Port:        realPort,
		BackendHost: realIP,
		BackendPort: realPort,
	}

	return route, nil
}
