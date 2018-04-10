package hardware

import (
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.DevicesReply {
	return &sonm.DevicesReply{
		CPU:     h.CPU,
		GPUs:    h.GPU,
		RAM:     h.RAM,
		Network: h.Network,
		Storage: h.Storage,
	}
}
