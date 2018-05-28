package sonm

import (
	"fmt"

	"github.com/cnf/structhash"
)

func (m *GPUDevice) FillHashID() {
	// reset prev hash value to not affecting the current hashing
	m.Hash = ""
	m.Hash = fmt.Sprintf("%x", structhash.Md5(*m, 1))
}

func (m *Network) Netflags() uint64 {
	return NetflagsToUint([3]bool{m.Overlay, m.Outbound, m.Incoming})
}
