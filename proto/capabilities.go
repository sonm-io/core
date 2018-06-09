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

const (
	MinNetFlagsCount = 3

	NetworkOverlay  = uint64(0x1)
	NetworkOutbound = uint64(0x2)
	NetworkIncoming = uint64(0x4)
)

func (m *NetFlags) ToBoolSlice() []bool {
	if m == nil {
		return make([]bool, MinNetFlagsCount)
	}
	result := []bool{}
	for idx := 0; idx < 64; idx++ {
		flag := uint64(1) << uint(idx)
		if uint64(flag) > m.Flags {
			break
		}
		flagValue := (m.Flags&flag == flag)
		result = append(result, flagValue)
	}
	for len(result) < MinNetFlagsCount {
		result = append(result, false)
	}
	return result
}

func NetFlagsFromBoolSlice(from []bool) *NetFlags {
	var flags uint64
	for idx, val := range from {
		if val {
			flags |= (1 << uint(idx))
		}
	}
	return &NetFlags{Flags: flags}
}

func (m *NetFlags) GetIncoming() bool {
	if m == nil {
		return false
	}
	return m.Flags&NetworkIncoming == NetworkIncoming
}

func (m *NetFlags) GetOutbound() bool {
	if m == nil {
		return false
	}
	return m.Flags&NetworkOutbound == NetworkOutbound
}

func (m *NetFlags) GetOverlay() bool {
	if m == nil {
		return false
	}
	return m.Flags&NetworkOverlay == NetworkOverlay
}

func (m *NetFlags) SetIncoming(value bool) *NetFlags {
	if value {
		m.Flags |= NetworkIncoming
	} else {
		m.Flags &= ^NetworkIncoming
	}
	return m
}

func (m *NetFlags) SetOutbound(value bool) *NetFlags {
	if value {
		m.Flags |= NetworkOutbound
	} else {
		m.Flags &= ^NetworkOutbound
	}
	return m
}

func (m *NetFlags) SetOverlay(value bool) *NetFlags {
	if value {
		m.Flags |= NetworkOverlay
	} else {
		m.Flags &= ^NetworkOverlay
	}
	return m
}

func (m *NetFlags) ConverseImplication(cmp *NetFlags) bool {
	return m.Flags|^cmp.GetFlags() == 1<<64-1
}
