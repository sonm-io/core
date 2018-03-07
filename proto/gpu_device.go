package sonm

import "github.com/cnf/structhash"

var Radeons = []uint64{
	4098,
	// macbook pro 2017
	16915456,
}

var Nvidias = []uint64{
	4318,
}

func TypeFromVendorID(v uint64) GPUVendorType {
	for _, id := range Radeons {
		if id == v {
			return GPUVendorType_RADEON
		}
	}

	for _, id := range Nvidias {
		if id == v {
			return GPUVendorType_NVIDIA
		}
	}

	return GPUVendorType_GPU_UNKNOWN
}

// VendorType returns GPU vendor type.
func (m *GPUDevice) VendorType() GPUVendorType {
	return TypeFromVendorID(m.VendorID)
}

func (m *GPUDevice) Hash() []byte {
	return structhash.Md5(m, 1)
}
